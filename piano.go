package main

import (
	"time"

	"github.com/schollz/portmidi"
	log "github.com/sirupsen/logrus"
)

// Piano is the AI class for the piano
type Piano struct {
	InputDevice  portmidi.DeviceID
	OutputDevice portmidi.DeviceID
	outputStream *portmidi.Stream
	inputStream  *portmidi.Stream
}

// Init sets the device ports. Optionally you can
// pass the input and output ports, respectively.
func (p *Piano) Init(ports ...int) (err error) {
	logger := log.WithFields(log.Fields{
		"function": "Piano.Init",
	})
	logger.Info("Initializing portmidi...")
	err = portmidi.Initialize()
	if err != nil {
		logger.WithFields(log.Fields{
			"msg": "initiailization failed",
		}).Error(err.Error())
		return
	}
	numDevices := portmidi.CountDevices()
	logger.Infof("Found %d devices", numDevices)
	for i := 0; i < numDevices; i++ {
		deviceInfo := portmidi.Info(portmidi.DeviceID(i))
		var inputOutput string
		if deviceInfo.IsOutputAvailable {
			inputOutput = "output"
			p.OutputDevice = portmidi.DeviceID(i)
		} else {
			inputOutput = "input"
			p.InputDevice = portmidi.DeviceID(i)
		}
		logger.Infof("%d) %s %s %s", i, deviceInfo.Interface, deviceInfo.Name, inputOutput)
	}
	if len(ports) == 2 {
		p.InputDevice = portmidi.DeviceID(ports[0])
		p.OutputDevice = portmidi.DeviceID(ports[1])
	}
	logger.Infof("Using input device %d and output device %d", p.InputDevice, p.OutputDevice)

	logger.Info("Opening output stream")
	p.outputStream, err = portmidi.NewOutputStream(p.OutputDevice, 1024, 0)
	if err != nil {
		if err != nil {
			logger.WithFields(log.Fields{
				"msg": "problem getting output stream from device " + string(p.OutputDevice),
			}).Error(err.Error())
			return
		}

	}

	logger.Info("Opening input stream")
	p.inputStream, err = portmidi.NewInputStream(p.InputDevice, 1024)
	if err != nil {
		if err != nil {
			logger.WithFields(log.Fields{
				"msg": "problem getting input stream from device " + string(p.InputDevice),
			}).Error(err.Error())
			return
		}

	}
	return
}

// Close will shutdown the streams
// and gracefully terminate.
func (p *Piano) Close() (err error) {
	logger := log.WithFields(log.Fields{
		"function": "Piano.Close",
	})
	logger.Debug("Closing output stream")
	p.outputStream.Close()
	logger.Debug("Closing input stream")
	p.inputStream.Close()
	logger.Debug("Terminating portmidi")
	portmidi.Terminate()
	return
}

// PlayNotes will execute a bunch of threads to play notes
func (p *Piano) PlayNotes(chord Chord, bpm float64) (err error) {
	for _, note := range chord.Notes {
		go p.PlayNote(note, bpm)
	}
	return
}

// PlayNote will play a single note. Hopefully this will work
// in a thread, but that remains to be seen (TODO).
// To turn on a note it will send 0x90 to the stream.
// To turn off a note it will send 0x80 to the stream.
func (p *Piano) PlayNote(note Note, bpm float64) (err error) {
	logger := log.WithFields(log.Fields{
		"function": "Piano.PlayNotes",
	})
	err = p.outputStream.WriteShort(0x90, note.Pitch, note.Velocity)
	if err != nil {
		logger.WithFields(log.Fields{
			"p": note.Pitch,
			"v": note.Velocity,
			"d": note.Duration,
		}).Error(err.Error())
	} else {
		logger.WithFields(log.Fields{
			"p": note.Pitch,
			"v": note.Velocity,
			"d": note.Duration,
		}).Info("on")
	}
	time.Sleep(time.Duration(note.Duration/bpm) * time.Minute)
	err = p.outputStream.WriteShort(0x80, note.Pitch, note.Velocity)
	if err != nil {
		logger.WithFields(log.Fields{
			"p": note.Pitch,
			"v": note.Velocity,
			"d": note.Duration,
		}).Error(err.Error())
	} else {
		logger.WithFields(log.Fields{
			"p": note.Pitch,
			"v": note.Velocity,
			"d": note.Duration,
		}).Info("off")
	}
	return
}