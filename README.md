# Alert Manager to AWS SNS

Forward Alert Manager alert to AWS SNS.

# How to use it

```bash
docker run -d --name am2sns -p 9876:9876 -e AWS_REGION=$AWS_REGION -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY -e AWS_SNS_TOPIC_ARN=$AWS_SNS_TOPIC_ARN scalair/am2sns
```

AWS user must have `SNS:Publish` permission.