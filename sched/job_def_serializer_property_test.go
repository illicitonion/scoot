// +build property_test

package sched

import (
	"github.com/scootdev/scoot/tests/testhelpers"
	"testing"
)

func Test_RandomSerializerDeserializer(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Serialize JobDef", prop.ForAll(
		func(job *Job) bool {

			if serializeBinary := ValidateSerialization(job, false); !serializeBinary {
				return false
			}

			if serializeJson := ValidateSerialization(job, true); !serializeJson {
				return false
			}

			return true
		},

		GopterGenJob(),
	))

	properties.TestingRun(t)
}

func GopterGenJob() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		numTasks := genParams.Rng.Intn(10)
		jobId := testhelpers.GenRandomAlphaNumericString(genParams.Rng)
		job := GenRandomJob(jobId, numTasks, genParams.Rng)

		genResult := gopter.NewGenResult(&job, gopter.NoShrinker)
		return genResult
	}
}

func GopterGenJobDef() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		jobId := testhelpers.GenRandomAlphaNumericString(genParams.Rng)
		numTasks := genParams.Rng.Intn(10)
		job := GenRandomJob(jobId, numTasks, genParams.Rng)
		genResult := gopter.NewGenResult(job.Def, gopter.NoShrinker)
		return genResult
	}
}
