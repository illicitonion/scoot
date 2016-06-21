package saga

import (
	"bytes"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"testing"
)

func Test_ValidateUpdateSagaState(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 1000
	properties := gopter.NewProperties(parameters)

	properties.Property("StartSaga message is never valid on an already started saga", prop.ForAll(
		func(state *SagaState, data []byte) bool {
			msg := MakeStartSagaMessage(state.SagaId(), data)
			newState, err := updateSagaState(state, msg)

			return err != nil && newState == nil
		},
		GenSagaState(),
		gen.SliceOf(gen.UInt8()),
	))

	// EndSaga messages are valid if a saga has not been Aborted and all StartTask have EndTask messages
	// If a saga has been aborted all StartTask messages must have corresponding StartCompTask / EndCompTask messages
	// for an EndSaga message to be valid.
	properties.Property("EndSaga message is Valid or returns an error", prop.ForAll(
		func(state *SagaState) bool {

			msg := MakeEndSagaMessage(state.SagaId())
			newState, err := updateSagaState(state, msg)

			validTransition := true
			for _, id := range state.GetTaskIds() {

				// if aborted all comp tasks must be completed for all tasks started
				// if not aborted all started tasks must be completed
				if state.IsSagaAborted() {
					if !(state.IsTaskStarted(id) &&
						state.IsCompTaskStarted(id) &&
						state.IsCompTaskCompleted(id)) {
						validTransition = false
					}
				} else {
					if !(state.IsTaskStarted(id) &&
						state.IsTaskCompleted(id)) {
						validTransition = false
					}
				}
			}

			// either we made a valid transition and had a valid update or applying
			// this message is an invalidTransition and an error was returned.
			validUpdate := validTransition && err == nil &&
				newState != nil && newState.IsSagaCompleted()
			errorReturned := !validTransition && err != nil && newState == nil

			return validUpdate || errorReturned
		},
		GenSagaState(),
	))

	// Abort messages are valid unless a Saga has been Completed
	properties.Property("AbortSaga message is valid or returns an error", prop.ForAll(
		func(state *SagaState) bool {
			validTransition := !state.IsSagaCompleted()

			msg := MakeAbortSagaMessage(state.SagaId())
			newState, err := updateSagaState(state, msg)

			// either we made a valid transition and had a valid update or applying
			// this message is an invalidTransition and an error was returned.
			validUpdate := validTransition && err == nil &&
				newState != nil && newState.IsSagaAborted()
			errorReturned := !validTransition && err != nil && newState == nil

			return validUpdate || errorReturned
		},
		GenSagaState(),
	))

	// StartTask messages are valid unless a Saga has been Completed or Aborted
	properties.Property("StartTask message is valid or returns an Error", prop.ForAll(

		func(pair StateTaskPair, data []byte) bool {

			state := &pair.state
			taskId := pair.taskId

			validTransition := !state.IsSagaCompleted() && !state.IsSagaAborted() && !state.IsTaskCompleted(taskId)

			msg := MakeStartTaskMessage(state.SagaId(), taskId, data)
			newState, err := updateSagaState(state, msg)

			// either we made a valid transition and had a valid update or applying
			// this message is an invalidTransition and an error was returned.
			validUpdate := validTransition && err == nil &&
				newState != nil && newState.IsTaskStarted(taskId) &&
				bytes.Equal(newState.GetStartTaskData(taskId), data)
			errorReturned := !validTransition && err != nil && newState == nil

			return validUpdate || errorReturned
		},
		GenSagaStateAndTaskId(),
		gen.SliceOf(gen.UInt8()),
	))

	// EndTask messages are valid if there is a corresponding StartTask message and a Saga
	// has not been aborted or completed
	properties.Property("EndTask message is valid or returns an Error", prop.ForAll(
		func(pair StateTaskPair, data []byte) bool {

			state := &pair.state
			taskId := pair.taskId

			validTransition := !state.IsSagaCompleted() && !state.IsSagaAborted() &&
				state.IsTaskStarted(taskId)

			msg := MakeEndTaskMessage(state.SagaId(), taskId, data)
			newState, err := updateSagaState(state, msg)

			// either we made a valid transition and had a valid update or applying
			// this message is an invalidTransition and an error was returned.
			validUpdate := validTransition && err == nil &&
				newState != nil && newState.IsTaskCompleted(taskId) &&
				bytes.Equal(newState.GetEndTaskData(taskId), data)
			errorReturned := !validTransition && err != nil && newState == nil

			return validUpdate || errorReturned
		},
		GenSagaStateAndTaskId(),
		gen.SliceOf(gen.UInt8()),
	))

	properties.Property("StartCompTask message is valid or returns an Error", prop.ForAll(
		func(pair StateTaskPair, data []byte) bool {

			state := &pair.state
			taskId := pair.taskId

			validTransition := state.IsSagaAborted() && !state.IsSagaCompleted() &&
				state.IsTaskStarted(taskId) && !state.IsCompTaskCompleted(taskId)

			msg := MakeStartCompTaskMessage(state.SagaId(), taskId, data)
			newState, err := updateSagaState(state, msg)

			// either we made a valid transition and had a valid update or applying
			// this message is an invalidTransition and an error was returned.
			validUpdate := validTransition && err == nil &&
				newState != nil && newState.IsCompTaskStarted(taskId) &&
				bytes.Equal(newState.GetStartCompTaskData(taskId), data)
			errorReturned := !validTransition && err != nil && newState == nil

			return validUpdate || errorReturned
		},
		GenSagaStateAndTaskId(),
		gen.SliceOf(gen.UInt8()),
	))

	properties.Property("EndCompTask message is valid or returns an Error", prop.ForAll(
		func(pair StateTaskPair, data []byte) bool {

			state := &pair.state
			taskId := pair.taskId

			validTransition := state.IsSagaAborted() && !state.IsSagaCompleted() &&
				state.IsTaskStarted(taskId) && state.IsCompTaskStarted(taskId)

			msg := MakeEndCompTaskMessage(state.SagaId(), taskId, data)
			newState, err := updateSagaState(state, msg)

			// either we made a valid transition and had a valid update or applying
			// this message is an invalidTransition and an error was returned.
			validUpdate := validTransition && err == nil &&
				newState != nil && newState.IsCompTaskCompleted(taskId) &&
				bytes.Equal(newState.GetEndCompTaskData(taskId), data)
			errorReturned := !validTransition && err != nil && newState == nil

			return validUpdate || errorReturned
		},
		GenSagaStateAndTaskId(),
		gen.SliceOf(gen.UInt8()),
	))

	properties.TestingRun(t)
}
