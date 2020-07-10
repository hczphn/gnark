// Copyright 2020 ConsenSys AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by gnark/internal/generators DO NOT EDIT

package groth16_test

import (
	curve "github.com/consensys/gurvy/bls381"
	"github.com/consensys/gurvy/bls381/fr"

	backend_bls381 "github.com/consensys/gnark/backend/bls381"
	"github.com/consensys/gnark/backend/r1cs"

	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"

	groth16_bls381 "github.com/consensys/gnark/backend/bls381/groth16"

	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/encoding/gob"
	"github.com/consensys/gurvy"
)

func TestCircuits(t *testing.T) {
	assert := groth16.NewAssert(t)
	matches, err := filepath.Glob("../../../internal/generators/testcircuits/generated/*.r1cs")

	if err != nil {
		t.Fatal(err)
	}

	if len(matches) == 0 {
		t.Fatal("couldn't find test circuits for", curve.ID.String())
	}
	for _, name := range matches {
		name = name[:len(name)-5]
		t.Log(curve.ID.String(), " -- ", filepath.Base(name))

		good := make(map[string]interface{})
		if err := gob.ReadMap(name+".good", good); err != nil {
			t.Fatal(err)
		}
		bad := make(map[string]interface{})
		if err := gob.ReadMap(name+".bad", bad); err != nil {
			t.Fatal(err)
		}
		var untypedR1CS r1cs.UntypedR1CS
		if err := gob.Read(name+".r1cs", &untypedR1CS, gurvy.UNKNOWN); err != nil {
			t.Fatal(err)
		}
		r1cs := untypedR1CS.ToR1CS(curve.ID)
		assert.NotSolved(r1cs, bad)
		assert.Solved(r1cs, good, nil)
	}
}

func TestParsePublicInput(t *testing.T) {

	expectedNames := [2]string{"data", backend.OneWire}

	inputOneWire := make(map[string]interface{})
	inputOneWire[backend.OneWire] = 3
	if _, err := groth16_bls381.ParsePublicInput(expectedNames[:], inputOneWire); err == nil {
		t.Fatal("expected ErrMissingAssigment error")
	}

	missingInput := make(map[string]interface{})
	if _, err := groth16_bls381.ParsePublicInput(expectedNames[:], missingInput); err == nil {
		t.Fatal("expected ErrMissingAssigment")
	}

	correctInput := make(map[string]interface{})
	correctInput["data"] = 3
	got, err := groth16_bls381.ParsePublicInput(expectedNames[:], correctInput)
	if err != nil {
		t.Fatal(err)
	}

	expected := make([]fr.Element, 2)
	expected[0].SetUint64(3).FromMont()
	expected[1].SetUint64(1).FromMont()
	if len(got) != len(expected) {
		t.Fatal("Unexpected length for assignment")
	}
	for i := 0; i < len(got); i++ {
		if !got[i].Equal(&expected[i]) {
			t.Fatal("error public assignment")
		}
	}

}

//--------------------//
//     benches		  //
//--------------------//

func referenceCircuit() (backend_bls381.R1CS, map[string]interface{}, map[string]interface{}) {

	name := "../../../../backend/groth16/testdata/" + strings.ToLower(curve.ID.String()) + "/reference_large"

	good := make(map[string]interface{})
	if err := gob.ReadMap(name+".good", good); err != nil {
		panic(err)
	}
	bad := make(map[string]interface{})
	if err := gob.ReadMap(name+".bad", bad); err != nil {
		panic(err)
	}

	var r1cs backend_bls381.R1CS

	if err := gob.Read(name+".r1cs", &r1cs, curve.ID); err != nil {
		panic(err)
	}

	return r1cs, good, bad
}

// BenchmarkSetup is a helper to benchmark Setup on a given circuit
func BenchmarkSetup(b *testing.B) {
	r1cs, _, _ := referenceCircuit()
	defer debug.SetGCPercent(debug.SetGCPercent(-1))
	var pk groth16_bls381.ProvingKey
	var vk groth16_bls381.VerifyingKey
	b.ResetTimer()

	b.Run("setup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			groth16_bls381.Setup(&r1cs, &pk, &vk)
		}
	})
}

// BenchmarkProver is a helper to benchmark Prove on a given circuit
// it will run the Setup, reset the benchmark timer and benchmark the prover
func BenchmarkProver(b *testing.B) {
	r1cs, solution, _ := referenceCircuit()
	defer debug.SetGCPercent(debug.SetGCPercent(-1))
	var pk groth16_bls381.ProvingKey
	var vk groth16_bls381.VerifyingKey
	groth16_bls381.Setup(&r1cs, &pk, &vk)

	b.ResetTimer()
	b.Run("prover", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = groth16_bls381.Prove(&r1cs, &pk, solution)
		}
	})
}

// BenchmarkVerifier is a helper to benchmark Verify on a given circuit
// it will run the Setup, the Prover and reset the benchmark timer and benchmark the verifier
// the provided solution will be filtered to keep only public inputs
func BenchmarkVerifier(b *testing.B) {
	r1cs, solution, _ := referenceCircuit()
	defer debug.SetGCPercent(debug.SetGCPercent(-1))
	var pk groth16_bls381.ProvingKey
	var vk groth16_bls381.VerifyingKey
	groth16_bls381.Setup(&r1cs, &pk, &vk)
	proof, err := groth16_bls381.Prove(&r1cs, &pk, solution)
	if err != nil {
		panic(err)
	}

	b.ResetTimer()
	b.Run("verifier", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = groth16_bls381.Verify(proof, &vk, solution)
		}
	})
}
