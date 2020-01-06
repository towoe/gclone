package repo

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

type configEncoder interface {
	Encode(w io.Writer, r *Register) error
}

type configDecoder interface {
	Decode(rj io.Reader, r *Register) error
}

type config struct {
	Encoder configEncoder
	Decoder configDecoder
}

func newConfigService() *config {
	return &config{
		Encoder: &jsonEncoder{},
		Decoder: &jsonDecoder{},
	}
}

func (c *config) Load(r *Register, storageFile string) error {
	rj, err := os.Open(storageFile)
	if err != nil {
		return err
	}
	defer rj.Close()
	c.Decoder.Decode(rj, r)
	return nil
}

func (c *config) Store(r *Register, storageFile string) error {
	w, err := os.OpenFile(storageFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		// TODO: better logging
		log.Println("storage file ", storageFile)
		return err
	}
	defer w.Close()

	return c.Encoder.Encode(w, r)
}

type jsonDecoder struct {
}

func (j *jsonDecoder) Decode(r io.Reader, c *Register) error {
	dec := json.NewDecoder(r)
	if err := dec.Decode(c); err != nil {
		return err
	}
	return nil
}

type jsonEncoder struct {
}

func (j *jsonEncoder) Encode(w io.Writer, c *Register) error {
	enc := json.NewEncoder(w)
	if err := enc.Encode(c); err != nil {
		return err
	}
	return nil
}
