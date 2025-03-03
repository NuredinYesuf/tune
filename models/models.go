// Package models defines the data structures used in the application.
//
// Couple represents a pairing of an anchor time and a song ID.
// AnchorTimeMs is the time in milliseconds that serves as a reference point.
// SongID is the identifier for the song.
//
// RecordData represents the metadata for an audio recording.
// Audio is the file path or URL to the audio file.
// Duration is the length of the audio in seconds.
// Channels is the number of audio channels (e.g., 2 for stereo).
// SampleRate is the number of samples per second (e.g., 44100 for CD quality).
// SampleSize is the number of bits per sample (e.g., 16 for CD quality).
package models

type Couple struct {
	AnchorTimeMs uint32
	SongID       uint32
}

type RecordData struct {
	Audio      string  `json:"audio"`
	Duration   float64 `json:"duration"`
	Channels   int     `json:"channels"`
	SampleRate int     `json:"sampleRate"`
	SampleSize int     `json:"sampleSize"`
}
