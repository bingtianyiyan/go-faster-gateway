//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
The MIT License (MIT)

Copyright (c) 2016-2020 Containous SAS; 2020-2025 Traefik Labs

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

// Code generated by deepcopy-gen. DO NOT EDIT.

package tls

import (
	types "go-faster-gateway/pkg/types"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertAndStores) DeepCopyInto(out *CertAndStores) {
	*out = *in
	out.Certificate = in.Certificate
	if in.Stores != nil {
		in, out := &in.Stores, &out.Stores
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertAndStores.
func (in *CertAndStores) DeepCopy() *CertAndStores {
	if in == nil {
		return nil
	}
	out := new(CertAndStores)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClientAuth) DeepCopyInto(out *ClientAuth) {
	*out = *in
	if in.CAFiles != nil {
		in, out := &in.CAFiles, &out.CAFiles
		*out = make([]types.FileOrContent, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClientAuth.
func (in *ClientAuth) DeepCopy() *ClientAuth {
	if in == nil {
		return nil
	}
	out := new(ClientAuth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GeneratedCert) DeepCopyInto(out *GeneratedCert) {
	*out = *in
	if in.Domain != nil {
		in, out := &in.Domain, &out.Domain
		*out = new(types.Domain)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GeneratedCert.
func (in *GeneratedCert) DeepCopy() *GeneratedCert {
	if in == nil {
		return nil
	}
	out := new(GeneratedCert)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Options) DeepCopyInto(out *Options) {
	*out = *in
	if in.CipherSuites != nil {
		in, out := &in.CipherSuites, &out.CipherSuites
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.CurvePreferences != nil {
		in, out := &in.CurvePreferences, &out.CurvePreferences
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	in.ClientAuth.DeepCopyInto(&out.ClientAuth)
	if in.ALPNProtocols != nil {
		in, out := &in.ALPNProtocols, &out.ALPNProtocols
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.PreferServerCipherSuites != nil {
		in, out := &in.PreferServerCipherSuites, &out.PreferServerCipherSuites
		*out = new(bool)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Options.
func (in *Options) DeepCopy() *Options {
	if in == nil {
		return nil
	}
	out := new(Options)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Store) DeepCopyInto(out *Store) {
	*out = *in
	if in.DefaultCertificate != nil {
		in, out := &in.DefaultCertificate, &out.DefaultCertificate
		*out = new(Certificate)
		**out = **in
	}
	if in.DefaultGeneratedCert != nil {
		in, out := &in.DefaultGeneratedCert, &out.DefaultGeneratedCert
		*out = new(GeneratedCert)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Store.
func (in *Store) DeepCopy() *Store {
	if in == nil {
		return nil
	}
	out := new(Store)
	in.DeepCopyInto(out)
	return out
}
