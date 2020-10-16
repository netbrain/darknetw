package multipart

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
)

type Option func(mw *multipart.Writer) error

func WithBoundary(boundary string) Option {
	return func(mw *multipart.Writer) error {
		return mw.SetBoundary(boundary)
	}
}

func WithPart(header textproto.MIMEHeader, buf []byte) Option {
	return WithPartFromReader(header, bytes.NewBuffer(buf))
}

func WithPartFromReader(header textproto.MIMEHeader, r io.Reader) Option {
	return func(mw *multipart.Writer) error {
		w, err := mw.CreatePart(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, r)
		return err
	}
}

func WithFormField(fieldname string, buf []byte) Option {
	return WithFormFieldFromReader(fieldname, bytes.NewBuffer(buf))
}

func WithFormFieldFromReader(fieldname string, r io.Reader) Option {
	return func(mw *multipart.Writer) error {
		w, err := mw.CreateFormField(fieldname)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, r)
		return err
	}
}

func WithFormFile(fieldname, file string) Option {
	return func(mw *multipart.Writer) error {
		fh, err := os.Open(file)
		if err != nil {
			return err
		}
		defer fh.Close()
		return WithFormFileFromReader(fieldname, filepath.Base(file), fh)(mw)
	}
}

func WithFormFileFromReader(fieldname, filename string, r io.Reader) Option {
	return func(mw *multipart.Writer) error {
		w, err := mw.CreateFormFile(fieldname, filename)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, r)
		return err
	}
}

func WriteMultipart(r *http.Request, writer io.Writer, options ...Option) error {
	mw := multipart.NewWriter(writer)
	defer mw.Close()
	for _, o := range options {
		err := o(mw)
		if err != nil {
			return err
		}
	}
	r.Header.Set("Content-Type", mw.FormDataContentType())
	return nil
}
