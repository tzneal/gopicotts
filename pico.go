package gopicotts

/*
#cgo LDFLAGS: -lttspico

#include <string.h>
#include <stdlib.h>
#include <picoapi.h>
#include <picoapid.h>
#include <picoos.h>

pico_Engine* cloneEngine(pico_Engine e) {
   pico_Engine* ret = calloc(sizeof(pico_Engine),1);
   memcpy(ret,&e,sizeof(pico_Engine));
   return ret;
}

picoos_SDFile* cloneFile(picoos_SDFile f) {
   picoos_SDFile* ret = calloc(sizeof(picoos_SDFile),1);
   memcpy(ret,&f,sizeof(picoos_SDFile));
   return ret;
}
pico_Resource* cloneResource(pico_Resource r) {
   pico_Resource* ret = calloc(sizeof(pico_Resource),1);
   memcpy(ret,&r,sizeof(pico_Resource));
   return ret;
}
pico_System* cloneSystem(pico_System s) {
   pico_System* ret = calloc(sizeof(pico_System),1);
   memcpy(ret,&s,sizeof(pico_System));
   return ret;
}
pico_Char* ptrAdd(pico_Char * buf, pico_Int16 x) {
  return buf + x;
}
*/
import "C"
import (
	"errors"
	"fmt"
	"path/filepath"
	"unsafe"
)

const picoVoiceName = "PicoVoice"

type Engine struct {
	// always initialized
	system             C.pico_System
	picoTaResource     C.pico_Resource
	picoTgResourceName string
	picoSgResource     C.pico_Resource
	picoSgResourceName string
	picoEngine         C.pico_Engine
	common             C.picoos_Common
	buf                []byte

	// initialized depending on mode
	outFile  C.picoos_SDFile
	outputFn func(data []int16)
}

type Options struct {
	Language    Language
	LanguageDir string
}

var DefaultOptions = Options{
	Language:    LanguageEnUS,
	LanguageDir: "/usr/share/pico/lang/",
}

// NewEngine creates a new Pico text to speech engine.
func NewEngine(opts Options) (*Engine, error) {
	const memSize = 2500000
	e := &Engine{buf: make([]byte, memSize)}
	ret := C.pico_initialize(unsafe.Pointer(&e.buf[0]), memSize, &e.system)
	if ret != 0 {
		return nil, e.getError("initializing pico system", ret)
	}
	if err := e.setLanguage(opts.Language, opts.LanguageDir); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *Engine) getError(desc string, ret C.pico_Status) error {
	if ret == 0 {
		return nil
	}
	buf := (*C.char)(C.calloc(C.PICO_RETSTRINGSIZE, 1))
	defer C.free(unsafe.Pointer(buf))
	C.pico_getSystemStatusMessage(e.system, ret, buf)
	return errors.New(fmt.Sprintf("error %s: %s", desc, C.GoString(buf)))
}

func (e *Engine) setLanguage(lang Language, path string) error {
	lInfo := getLangInfo(lang)

	// Load the text analysis Lingware resource file
	langFile := filepath.Join(path, lInfo.internalTaLingware)
	fn := C.CString(langFile)
	defer C.free(unsafe.Pointer(fn))

	ret := C.pico_loadResource(e.system, (*C.pico_Char)(unsafe.Pointer(fn)), &e.picoTaResource)
	if ret != 0 {
		return e.getError("loading text analysis", ret)
	}

	// Load the signal generation Lingware resource file
	langFile = filepath.Join(path, lInfo.internalSgLingware)
	fn = C.CString(langFile)
	defer C.free(unsafe.Pointer(fn))
	ret = C.pico_loadResource(e.system, (*C.pico_Char)(unsafe.Pointer(fn)), &e.picoSgResource)
	if ret != 0 {
		return e.getError("loading signal analysis", ret)
	}

	// Get the text analysis resource name.
	ptrn := (*C.char)(C.calloc(C.PICO_MAX_RESOURCE_NAME_SIZE, 1))
	defer C.free(unsafe.Pointer(ptrn))
	ret = C.pico_getResourceName(e.system, e.picoTaResource, ptrn)
	if ret != 0 {
		return e.getError("determining text analysis resource name", ret)
	}
	e.picoTgResourceName = C.GoString(ptrn)

	// Get the signal generation resource name
	sgrn := (*C.char)(C.calloc(C.PICO_MAX_RESOURCE_NAME_SIZE, 1))
	defer C.free(unsafe.Pointer(sgrn))
	ret = C.pico_getResourceName(e.system, e.picoSgResource, sgrn)
	if ret != 0 {
		return e.getError("determining signal analysis resource name", ret)
	}
	e.picoSgResourceName = C.GoString(sgrn)

	// Create a voice definition
	vn := C.CString(picoVoiceName)
	defer C.free(unsafe.Pointer(vn))
	ret = C.pico_createVoiceDefinition(e.system, (*C.pico_Char)(unsafe.Pointer(vn)))
	if ret != 0 {
		return e.getError("creating voice definition", ret)
	}

	// Add the text analysis resource to the voice.
	pvn := C.CString(picoVoiceName)
	defer C.free(unsafe.Pointer(pvn))
	ret = C.pico_addResourceToVoiceDefinition(e.system, (*C.pico_Char)(unsafe.Pointer(pvn)), (*C.pico_Char)(unsafe.Pointer(ptrn)))
	if ret != 0 {
		return e.getError("assigning text analysis resource to voice", ret)
	}

	// Add the signal generation resource to the voice.
	ret = C.pico_addResourceToVoiceDefinition(e.system, (*C.pico_Char)(unsafe.Pointer(pvn)), (*C.pico_Char)(unsafe.Pointer(sgrn)))
	if ret != 0 {
		return e.getError("assigning signal generation resource to voice", ret)
	}

	// Create a new Pico engine.
	ret = C.pico_newEngine(e.system, (*C.pico_Char)(unsafe.Pointer(pvn)), &e.picoEngine)
	if ret != 0 {
		return e.getError("creating pico engine", ret)
	}

	e.common = C.pico_sysGetCommon(e.system)
	return nil
}

// SetFileOutput sets the output file for encoded speech.  If the
// filename ends in .wav, it will be WAVE formatted, if it ends in
// .au, it will be .AU formatted, otherwise it will be raw headerless
// PCM audio.
func (e *Engine) SetFileOutput(filename string) error {
	cfn := C.CString(filename)
	defer C.free(unsafe.Pointer(cfn))
	if C.picoos_sdfOpenOut(e.common, &e.outFile, (*C.picoos_char)(unsafe.Pointer(cfn)), C.SAMPLE_FREQ_16KHZ, C.PICOOS_ENC_LIN) != C.TRUE {
		return fmt.Errorf("unable to open output: %s", filename)
	}
	return nil
}

func (e *Engine) CloseFileOutput() error {
	// avoiding a go pointer to go pointer by cloning the outfile for
	// sdfCloseOut, which we then no longer need
	of := C.cloneFile(e.outFile)
	defer C.free(unsafe.Pointer(of))
	ret := C.picoos_sdfCloseOut(e.common, of)
	e.outFile = nil

	if ret != C.TRUE {
		return errors.New("closing output file")
	}
	return nil
}

// If fn is non-nil, it will be called with raw 16000hz mono generated audio
func (e *Engine) SetOutput(fn func(data []int16)) {
	e.outputFn = fn
}

func (e *Engine) deliverOutput(buf []int16) {
	if e.outFile != nil {
		C.picoos_sdfPutSamples(e.outFile, (C.picoos_uint32)(len(buf)), (*C.picoos_int16)(unsafe.Pointer(&buf[0])))
	}
	if e.outputFn != nil {
		e.outputFn(buf)
	}
}

func (e *Engine) processOutput() error {
	const chunkSize = 128
	const bufSize = chunkSize * 16
	var out_data_type, bytes_recv C.pico_Int16
	buf := make([]int16, bufSize)

	// Retrieve the samples and add them to the buffer
	done := false
	offset := 0 // offset in samples (2 bytes per)
	for !done {
		// might read too much data next time, so deliver what we've got now
		if offset+chunkSize/2 >= bufSize {
			e.deliverOutput(buf[0:offset])
			offset = 0
		}
		ret := C.pico_getData(e.picoEngine, unsafe.Pointer(&buf[offset]), bufSize, &bytes_recv, &out_data_type)
		// samples are 16 bit
		samplesRecived := int(bytes_recv) / 2
		offset += samplesRecived

		switch ret {
		case C.PICO_STEP_BUSY: // still processing
		case C.PICO_STEP_IDLE: // completed
			done = true
		default:
			return e.getError("getting data", ret)
		}

		// no data
		if samplesRecived == 0 {
			continue
		}

	}
	if offset != 0 {
		e.deliverOutput(buf[0:offset])
	}
	return nil
}

// FlushSendText flushes the Pico input buffer, which might be waiting
// on an end of sentence before producing speech.
func (e *Engine) FlushSendText() error {
	buf := (*C.char)(C.calloc(1, 1))
	defer C.free(unsafe.Pointer(buf))
	var sent C.pico_Int16
	ret := C.pico_putTextUtf8(e.picoEngine, (*C.pico_Char)(unsafe.Pointer(buf)), (C.pico_Int16)(1), &sent)

	if ret != 0 {
		return e.getError("sending text", ret)
	}

	if err := e.processOutput(); err != nil {
		return err
	}
	return nil
}

// SendText sends 'text' encoded in UTF8 into the Pico text input
// buffer. The input text may also contain text-input commands to
// change, for example, speed or pitch of the resulting speech output.
// Sentence ends are automatically detected.
func (e *Engine) SendText(s string) error {
	remaining := (C.pico_Int16)(len(s))
	buf := C.CString(s)
	defer C.free(unsafe.Pointer(buf))
	ptr := (*C.pico_Char)(unsafe.Pointer(buf))

	for remaining > 0 {
		var sent C.pico_Int16
		ret := C.pico_putTextUtf8(e.picoEngine, ptr, remaining, &sent)
		if ret != 0 {
			return e.getError("sending text", ret)
		}

		remaining -= sent
		ptr = C.ptrAdd(ptr, sent)
		if err := e.processOutput(); err != nil {
			return err
		}
	}
	return nil
}

// Close shuts down the engine and frees resources
func (e *Engine) Close() error {
	pvn := C.CString(picoVoiceName)
	defer C.free(unsafe.Pointer(pvn))
	pvnc := (*C.pico_Char)(unsafe.Pointer(pvn))

	if e.picoEngine != nil {
		eng := C.cloneEngine(e.picoEngine)
		defer C.free(unsafe.Pointer(eng))
		C.pico_disposeEngine(e.system, eng)
		e.picoEngine = nil
	}

	C.pico_releaseVoiceDefinition(e.system, pvnc)

	if e.picoSgResource != nil {
		sg := C.cloneResource(e.picoSgResource)
		defer C.free(unsafe.Pointer(sg))
		C.pico_unloadResource(e.system, sg)
	}

	if e.picoTaResource != nil {
		ta := C.cloneResource(e.picoTaResource)
		defer C.free(unsafe.Pointer(ta))
		C.pico_unloadResource(e.system, ta)
	}

	sys := C.cloneSystem(e.system)
	defer C.free(unsafe.Pointer(sys))

	C.pico_terminate(sys)
	return nil
}
