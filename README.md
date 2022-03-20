# Cloudformation Update Stack Github action

Welcome. This is a simple github action to make updating cloudformation stacks
a little less painful.

This is not a fancy tool, it just tries to do the right thing for the way I tend
to use cloudformation - patches and issues are welcome, but this is unlikely to
ever be extensively expanded.

## Inputs

### `stack-id`

**required** The Stack name or ARN that the update operation will be performed on.

### `template-file`

The path to the file containing the template (may be relative)

If neither `template-file` or `template-url` are specified, the current template is re-used
 
### `template-url`

The URL containing the template (must be an S3 URL)

If neither `template-file` or `template-url` are specified, the current template is re-used

### `wait-before`

If `true` wait for the stack to be in a "clean" state before running update.
This will consume runner minutes, but may help to reduce failures if you have
frequent updates or multiple people committing.

### `wait-after`

Same as wait-before, but will wait after the update.  If the update fails,
the build will fail.

### `parameter-*`

For any string after the `parameter-` prefix, it will be matched (case-insenitively)
to a parameter on the stack, and will update the parameter.

Any parameter omitted from this block will be left at the current value.

## Notes

### IAM Permissions

The de-minimus role required will look something like:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "UpdateSpecificStack",
            "Effect": "Allow",
            "Action": [
                "cloudformation:DescribeStacks",
                "cloudformation:UpdateStack"
            ],
            "Resource": "arn:aws:cloudformation:A-REGION:123456789012:stack/MyStackName/*"
        },
    ]
}
```

This would allow the tool to update a single specific stack (MyStackName), but note it will be
unable to replace the template itself -- this is an implementation detail not an IAM limitation.

You can of course broaden the ARN as appropriate to your environment.

To enable updating the stack, you must add a second statement similar to:

```json
        {
            "Sid": "ValidateTemplate",
            "Effect": "Allow",
            "Action": [
                "cloudformation:ValidateTemplate"
            ],
            "Resource": "*"
        }
```

This is because in a stack replacement scenario, it needs to know what Paramaters are still
valid (or newly created) so that it can properly write the correct update.

Finally, it is _recommended_ that your stack be deployed with a cloudformation role account that
will perform the underlying updates, otherwise your role will need permissions to read/mutate/delete
any resources your stack manipulates.

### Capabilities

Capabilities are either carried forward from the running stack (assuming no template changes),
or retrieved from the Validation stage - it cannot at this time be configured by the user.

## Example usage

```yaml
- name: Update Stack
  uses: jdnurmi/cfn-update-action@master
  with:
    stack-id: MyStackName
    wait-before: true
    wait-after: true
    parameter-RepoSha: ${{ github.sha }}
    template-file: ./mytpl.cfn
```


