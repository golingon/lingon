// Package sylt implements a workflow and actions engine for lingon.
//
// Actions can be implemented for specific tooling by implementing the
// [Actioner] interface.
// For example, [TerraAction] exists for deploying terra stacks.
//
// A [Workflow] type exists to combine and multiple actions into a workflow.
// Actions should be designed to be runnable independently of a workflow.
//
// "Sylt" is a Swedish word for jam.
// This package is where we turn lingon into sylt.
package sylt
