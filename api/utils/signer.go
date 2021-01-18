/*
Copyright 2021 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"io"

	"golang.org/x/crypto/ssh"
)

// AlgSigner wraps a provided ssh.Signer to ensure signature algorithm
// compatibility with OpenSSH.
//
// Right now it allows forcing SHA-2 signatures with RSA keys, instead of the
// default SHA-1 used by x/crypto/ssh. See
// https://www.openssh.com/txt/release-8.2 for context.
//
// If the provided Signer is not an RSA key or does not implement
// ssh.AlgorithmSigner, it's returned as is.
//
// DELETE IN 5.0: assuming https://github.com/golang/go/issues/37278 is fixed
// by then and we pull in the fix. Also delete all call sites.
func AlgSigner(s ssh.Signer, alg string) ssh.Signer {
	if alg == "" {
		return s
	}
	if s.PublicKey().Type() != ssh.KeyAlgoRSA && s.PublicKey().Type() != ssh.CertAlgoRSAv01 {
		return s
	}
	as, ok := s.(ssh.AlgorithmSigner)
	if !ok {
		return s
	}
	return fixedAlgorithmSigner{
		AlgorithmSigner: as,
		alg:             alg,
	}
}

type fixedAlgorithmSigner struct {
	ssh.AlgorithmSigner
	alg string
}

func (s fixedAlgorithmSigner) SignWithAlgorithm(rand io.Reader, data []byte, alg string) (*ssh.Signature, error) {
	if alg == "" {
		alg = s.alg
	}
	return s.AlgorithmSigner.SignWithAlgorithm(rand, data, alg)
}

func (s fixedAlgorithmSigner) Sign(rand io.Reader, data []byte) (*ssh.Signature, error) {
	return s.AlgorithmSigner.SignWithAlgorithm(rand, data, s.alg)
}
