package belvedere

import (
	"context"
	"fmt"
	"time"

	"go.opencensus.io/trace"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/serviceusage/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	requiredServices = []string{
		"cloudbuild.googleapis.com",
		"clouddebugger.googleapis.com",
		"clouderrorreporting.googleapis.com",
		"cloudkms.googleapis.com",
		"cloudprofiler.googleapis.com",
		"cloudresourcemanager.googleapis.com",
		"cloudtrace.googleapis.com",
		"compute.googleapis.com",
		"containeranalysis.googleapis.com",
		"containerregistry.googleapis.com",
		"containerscanning.googleapis.com",
		"deploymentmanager.googleapis.com",
		"dns.googleapis.com",
		"iam.googleapis.com",
		"iap.googleapis.com",
		"logging.googleapis.com",
		"monitoring.googleapis.com",
		"oslogin.googleapis.com",
		"stackdriver.googleapis.com",
		"storage-api.googleapis.com",
	}
)

// EnableServices enables all required services for the given GCP project.
func EnableServices(ctx context.Context, projectID string) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.EnableServices")
	defer span.End()

	su, err := serviceusage.NewService(ctx)
	if err != nil {
		return err
	}

	// Divide the required services up into batches of at most 20 services.
	for _, serviceIDs := range batchStrings(requiredServices, 20) {
		// Enable the services.
		op, err := su.Services.BatchEnable(
			fmt.Sprintf("projects/%s", projectID),
			&serviceusage.BatchEnableServicesRequest{
				ServiceIds: serviceIDs,
			},
		).Do()
		if err != nil {
			return err
		}

		// Record which services we enabled.
		for _, service := range serviceIDs {
			span.Annotate([]trace.Attribute{
				trace.StringAttribute("service", service),
				trace.StringAttribute("operation", op.Name),
			}, "Service enabled")
		}

		// Wait for the services to be enabled.
		f := checkServiceUsageOperation(ctx, su, op)
		if err := wait.Poll(10*time.Second, 5*time.Minute, f); err != nil {
			return err
		}
	}

	return nil
}

// EnableDeploymentManagerIAM binds the Deployment Manager service account to the `owner` role if
// it has not already been so bound. This allows Deployment Manager to add IAM roles to service
// accounts per https://cloud.google.com/deployment-manager/docs/configuration/set-access-control-resources#granting_deployment_manager_permission_to_set_iam_policies
func EnableDeploymentManagerIAM(ctx context.Context, projectID string) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.EnableDeploymentManagerIAM")
	defer span.End()

	crm, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return err
	}

	// Resolve the project's numeric ID.
	project, err := crm.Projects.Get(projectID).Fields("projectNumber").Do()
	if err != nil {
		return err
	}

	// Get the project's IAM policy.
	policy, err := crm.Projects.GetIamPolicy(projectID, &cloudresourcemanager.GetIamPolicyRequest{}).Do()
	if err != nil {
		return err
	}

	crmMember := fmt.Sprintf("serviceAccount:%d@cloudservices.gserviceaccount.com", project.ProjectNumber)
	const owner = "roles/owner"

	// Look for an existing IAM binding giving Deployment Manager ownership of the project.
	for _, binding := range policy.Bindings {
		if binding.Role == owner {
			for _, member := range binding.Members {
				if member == crmMember {
					span.Annotate(
						[]trace.Attribute{
							trace.Int64Attribute("project_id", project.ProjectNumber),
						},
						"Binding verified",
					)
					return nil
				}
			}
		}
	}

	// If none exists, add a binding and update the policy.
	span.Annotate(
		[]trace.Attribute{
			trace.Int64Attribute("project_id", project.ProjectNumber),
		},
		"Binding created",
	)
	policy.Bindings = append(policy.Bindings, &cloudresourcemanager.Binding{
		Members: []string{crmMember},
		Role:    owner,
	})
	_, err = crm.Projects.SetIamPolicy(projectID, &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}).Do()
	return err
}

func checkServiceUsageOperation(ctx context.Context, su *serviceusage.Service, op *serviceusage.Operation) wait.ConditionFunc {
	return func() (bool, error) {
		_, span := trace.StartSpan(ctx, "belvedere.checkServiceUsageOperation")
		defer span.End()

		o, err := su.Operations.Get(op.Name).Do()
		if err != nil {
			return false, err
		}

		if o.Error != nil {
			span.Annotate([]trace.Attribute{
				trace.Int64Attribute("error.code", o.Error.Code),
				trace.StringAttribute("error.message", o.Error.Message),
				trace.StringAttribute("error.details", fmt.Sprint(o.Error.Details)),
			}, "Error")
			span.SetStatus(trace.Status{Code: trace.StatusCodeAborted, Message: o.Error.Message})
		}
		span.AddAttributes(trace.BoolAttribute("done", op.Done))
		return o.Done, nil
	}
}
