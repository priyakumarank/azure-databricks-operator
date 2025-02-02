/*
Copyright 2019 microsoft.

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

package controllers

import (
	"context"
	"fmt"

	databricksv1 "github.com/microsoft/azure-databricks-operator/api/v1"
)

const finalizerName = "notebookjob.finalizers.databricks.microsoft.com"

func (r *NotebookJobReconciler) addFinalizer(instance *databricksv1.NotebookJob) error {
	instance.AddFinalizer(finalizerName)
	err := r.Update(context.Background(), instance)
	if err != nil {
		return fmt.Errorf("failed to update finalizer: %v", err)
	}
	r.Recorder.Event(instance, "Normal", "Updated", fmt.Sprintf("finalizer %s added", finalizerName))
	return nil
}

func (r *NotebookJobReconciler) handleFinalizer(instance *databricksv1.NotebookJob) error {
	if instance.HasFinalizer(finalizerName) {
		// our finalizer is present, so lets handle our external dependency
		if err := r.deleteExternalDependency(instance); err != nil {
			return err
		}

		instance.RemoveFinalizer(finalizerName)
		if err := r.Update(context.Background(), instance); err != nil {
			return err
		}
	}
	// Our finalizer has finished, so the reconciler can do nothing.
	return nil
}

func (r *NotebookJobReconciler) deleteExternalDependency(instance *databricksv1.NotebookJob) error {
	if instance.Status.Run != nil {
		r.Log.Info(fmt.Sprintf("Deleting external dependencies (run_id: %d)", instance.Status.Run.RunID))
	}
	return r.deleteRunFromDatabricks(instance)
}
