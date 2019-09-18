package main

import (
	"fmt"
	//Workflow variables
	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func processBuckets(sBucket wfv1.Artifact,
	path string,
	endpoint string,
	bucketName string,
	key string,
	accessKey string,
	aredentialSecretName string,
	aecretKey string) wfv1.Artifact {

	sBucket.Path = path
	sBucket.ArtifactLocation.S3.Endpoint = endpoint
	sBucket.ArtifactLocation.S3.Bucket = bucketName
	sBucket.ArtifactLocation.S3.Key = key
	sBucket.ArtifactLocation.S3.AccessKeySecret.Key = accessKey
	sBucket.ArtifactLocation.S3.AccessKeySecret.Name = aredentialSecretName
	sBucket.ArtifactLocation.S3.SecretKeySecret.Key = aecretKey
	sBucket.ArtifactLocation.S3.SecretKeySecret.Name = aredentialSecretName

	return sBucket
}

func processEnv(config WorkflowConfig, workflowStep int) []apiv1.EnvVar {
	envs := make([]apiv1.EnvVar, len(config.WorkflowSteps[workflowStep].Environments)+3)
	for i, env := range config.WorkflowSteps[workflowStep].Environments {
		var newEnv apiv1.EnvVar
		newEnv.Name = env.Name
		newEnv.Value = env.Value
		envs[i] = newEnv
	}

	var inBucketEnv apiv1.EnvVar
	inBucketEnv.Name = "IN_FOLDER"
	inBucketEnv.Value = config.WorkflowSteps[workflowStep].InputBucket.BucketPath
	envs[len(config.WorkflowSteps[workflowStep].Environments)] = inBucketEnv

	var outBucketEnv apiv1.EnvVar
	outBucketEnv.Name = "OUT_FOLDER"
	outBucketEnv.Value = config.WorkflowSteps[workflowStep].OutputBucket.BucketPath
	envs[len(config.WorkflowSteps[workflowStep].Environments)+1] = outBucketEnv

	var queueName apiv1.EnvVar
	queueName.Name = "QUEUE_NAME"
	queueName.Value = config.WorkflowSteps[workflowStep].WorkQueueInfo.QueueName
	envs[len(config.WorkflowSteps[workflowStep].Environments)+2] = queueName

	return envs
}

func setResources(config WorkflowConfig) apiv1.ResourceRequirements {
	var resources apiv1.ResourceRequirements

	resourcelist := make(map[apiv1.ResourceName]resource.Quantity)
	resourcelist[apiv1.ResourceCPU] = resource.MustParse(fmt.Sprintf("%v", config.ResourcesRequest.Cpu))
	resourcelist[apiv1.ResourceMemory] = resource.MustParse(config.ResourcesRequest.Memory)

	resources.Requests = resourcelist

	return resources
}

func createWorkFlowTemplate(workflow wfv1.Workflow, config WorkflowConfig, workflowTemplateName string) wfv1.Template {
	workflowTemplate := workflow.Spec.Templates[0].DeepCopy() //Defines the workflow
	workflowTemplate.Name = workflowTemplateName
	return *workflowTemplate
}

func createParallelWorkflowSteps(config WorkflowConfig, workflowTemplateName string) [][]wfv1.WorkflowStep{
	wfStep := make([][]wfv1.WorkflowStep, len(config.WorkflowSteps))
	for stepNumber := range wfStep {
		wfStep[stepNumber] = make([]wfv1.WorkflowStep, 1)

		step := &wfStep[stepNumber][0] //Use Alias pointer to simplify code.
		step.Name = fmt.Sprintf("%s%d","Training", stepNumber)
		step.Template = fmt.Sprintf("%s%d",workflowTemplateName, stepNumber)

		//Create number of items per stem. One item is one "worker" or parallel step
		items := make([]wfv1.Item, config.WorkflowSteps[stepNumber].NumOfWorkers)
		//Set worker names here.
		for workerIndex := 0; workerIndex < config.WorkflowSteps[stepNumber].NumOfWorkers; workerIndex++ {
			items[workerIndex].Value = ("-Worker-" + fmt.Sprint(workerIndex))
		}

		step.WithItems = items
	}

	return wfStep
}

func createParallelDefinitionTemplate(workflow wfv1.Workflow, config WorkflowConfig, workflowTemplateName string, parallelTemplateName string) wfv1.Template {
	parallelTemplate := wfv1.Template{} //Defines the parallel steps

	//create the new template
	parallelTemplate.Name = parallelTemplateName

	//Define the parallel machines
	parallelTemplate.Steps = createParallelWorkflowSteps(config, workflowTemplateName)

	return parallelTemplate
}

func setInstanceTypeLabel(config WorkflowConfig) map[string]string {
	if (config.ResourcesRequest.InstanceType != "") {
		labels := make(map[string]string)
		labels["InstanceType"] = config.ResourcesRequest.InstanceType
		fmt.Println("SETTING INSTANCE TYPE: ", config.ResourcesRequest.InstanceType)
		return labels
	}

	return nil
}

func processWorkflow(workflow wfv1.Workflow, config WorkflowConfig) wfv1.Workflow {
	workflowTemplateName := "WorkflowTraining"
	parallelTemplateName := "Parallel-Test"

	//Create the new template list
	templates := make([]wfv1.Template, len(config.WorkflowSteps) + 1)
	templates[len(config.WorkflowSteps)] = createParallelDefinitionTemplate(workflow,config,workflowTemplateName, parallelTemplateName) //FIX THE NAME!!
	for workflowStepNumber, workflowStep := range config.WorkflowSteps {
		templates[workflowStepNumber] = createWorkFlowTemplate(workflow,config, fmt.Sprintf("%s%d",workflowTemplateName, workflowStepNumber))

		//Alias for the template defining the main workflow
		workflowTemplate := templates[workflowStepNumber]

		workflowTemplate.Container.Image = workflowStep.WorkflowImage
		workflowTemplate.Container.Env = processEnv(config, workflowStepNumber)

		workflowTemplate.Metadata.Labels = setInstanceTypeLabel(config)


		if config.ResourcesRequest.Request {
			workflowTemplate.Container.Resources = setResources(config)
		}

		if workflowStep.InputBucket.Enabled {
			workflowTemplate.Inputs.Artifacts[0] = processBuckets(
				workflowTemplate.Inputs.Artifacts[0],
				workflowStep.InputBucket.BucketPath,
				workflowStep.InputBucket.Endpoint,
				workflowStep.InputBucket.BucketName,
				workflowStep.InputBucket.Key,
				workflowStep.InputBucket.AccessKey,
				workflowStep.InputBucket.CredentialSecretName,
				workflowStep.InputBucket.SecretKey)
		} else {
			workflowTemplate.Inputs.Artifacts = nil
		}

		if workflowStep.OutputBucket.Enabled {
			workflowTemplate.Outputs.Artifacts[0] = processBuckets(
				workflowTemplate.Outputs.Artifacts[0],
				workflowStep.OutputBucket.BucketPath,
				workflowStep.OutputBucket.Endpoint,
				workflowStep.OutputBucket.BucketName,
				workflowStep.OutputBucket.Key,
				workflowStep.OutputBucket.AccessKey,
				workflowStep.OutputBucket.CredentialSecretName,
				workflowStep.OutputBucket.SecretKey)
		} else {
			workflowTemplate.Outputs.Artifacts = nil
		}


		templates[workflowStepNumber] = workflowTemplate
	}

	//Set the correct entry and replace the old template with the new one
	workflow.Spec.Entrypoint = templates[len(config.WorkflowSteps)].Name
	workflow.Spec.Templates = templates

	return workflow
}

