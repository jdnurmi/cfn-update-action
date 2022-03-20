package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

func EnvMap(prefix string) (env map[string]string) {
	env = map[string]string{}
	for _, item := range os.Environ() {
		kv := strings.SplitN(item, "=", 2)
		if len(kv) != 2 {
			log.Fatalf("Bad environment variable: %+v", kv)
		}
		if strings.HasPrefix(kv[0], prefix) {
			env[kv[0][len(prefix):]] = kv[1]
		}
	}
	return
}

func main() {
	var (
		TemplateFile    = os.Getenv("INPUT_TEMPLATE-FILE")
		TemplateBody    string
		TemplateURL     = os.Getenv("INPUT_TEMPLATE-URL")
		StackId         = os.Getenv("INPUT_STACK-ID")
		InputParameters = EnvMap("INPUT_PARAMETER-")
		WaitBefore      = os.Getenv("INPUT_WAIT-BEFORE") == "true"
		WaitAfter       = os.Getenv("INPUT_WAIT-AFTER") == "true"
		Parameters      = map[string]*types.Parameter{}
		Capabilities    []types.Capability
	)
	awsCfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Couldn't configure aws: %v", err)
	}
	cfn := cloudformation.NewFromConfig(awsCfg)
	if TemplateFile != "" {
		body, err := ioutil.ReadFile(TemplateFile)
		if err != nil {
			log.Fatalf("Failed to load template from %q: %v", TemplateFile, err)
		}
		TemplateBody = string(body)
		validation, err := cfn.ValidateTemplate(context.TODO(), &cloudformation.ValidateTemplateInput{
			TemplateBody: &TemplateBody,
		})
		if err != nil {
			log.Fatalf("Failed to validate template %q: %v", TemplateURL, err)
		}
		Capabilities = validation.Capabilities
		for _, parameter := range validation.Parameters {
			Parameters[strings.ToUpper(*parameter.ParameterKey)] = &types.Parameter{
				ParameterKey:     parameter.ParameterKey,
				UsePreviousValue: aws.Bool(true),
			}
		}
	} else if TemplateURL != "" {
		validation, err := cfn.ValidateTemplate(context.TODO(), &cloudformation.ValidateTemplateInput{
			TemplateURL: &TemplateURL,
		})
		if err != nil {
			log.Fatalf("Failed to validate template %q: %v", TemplateURL, err)
		}
		Capabilities = validation.Capabilities
		for _, parameter := range validation.Parameters {
			Parameters[strings.ToUpper(*parameter.ParameterKey)] = &types.Parameter{
				ParameterKey:     parameter.ParameterKey,
				UsePreviousValue: aws.Bool(true),
			}
		}
	} else {
		description, err := cfn.DescribeStacks(context.TODO(), &cloudformation.DescribeStacksInput{
			StackName: &StackId,
		})
		if err != nil {
			log.Fatalf("Failed to describe stack %q: %v", StackId, err)
		}
		if len(description.Stacks) != 1 {
			log.Fatalf("Stack %q was not returned", StackId)
		}
		Capabilities = description.Stacks[0].Capabilities
		for _, parameter := range description.Stacks[0].Parameters {
			// Assume we will keep the value unless we get an explicit update
			Parameters[strings.ToUpper(*parameter.ParameterKey)] = &types.Parameter{
				ParameterKey:     parameter.ParameterKey,
				UsePreviousValue: aws.Bool(true),
			}
		}
	}

	for k, v := range InputParameters {
		// The parameter exists, so we are replacing it
		if _, ok := Parameters[k]; ok {
			Parameters[k].UsePreviousValue = aws.Bool(false)
			Parameters[k].ParameterValue = aws.String(v)
		} else {
			// The parameter doesn't exist and isn't defined
			log.Fatalf("Input parameter %q is not known to the stack", k)
		}
	}

	inp := &cloudformation.UpdateStackInput{
		StackName:    &StackId,
		Capabilities: Capabilities,
	}

	inp.Parameters = make([]types.Parameter, 0, len(Parameters))
	for _, p := range Parameters {
		inp.Parameters = append(inp.Parameters, *p)
	}
	if TemplateBody != "" {
		inp.TemplateBody = &TemplateBody
	} else if TemplateURL != "" {
		inp.TemplateURL = &TemplateURL
	} else {
		inp.UsePreviousTemplate = aws.Bool(true)
	}

	if WaitBefore {
		log.Printf("Waiting for stack to be stable before updating")
		err = cloudformation.NewStackUpdateCompleteWaiter(cfn).Wait(context.TODO(), &cloudformation.DescribeStacksInput{
			StackName: &StackId,
		}, time.Minute*60)
		if err != nil {
			log.Fatalf("Wait before update failed: %v", err)
		}
	}
	_, err = cfn.UpdateStack(context.TODO(), inp)
	if err != nil {
		log.Fatalf("Couldn't update stack: %v", err)
	}
	if WaitAfter {
		log.Printf("Waiting for stack to be stable after update")
		err = cloudformation.NewStackUpdateCompleteWaiter(cfn).Wait(context.TODO(), &cloudformation.DescribeStacksInput{
			StackName: &StackId,
		}, time.Minute*60)
		if err != nil {
			log.Fatalf("Wait before update failed: %v", err)
		}
	}

}
