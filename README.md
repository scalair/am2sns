# Alert Manager to AWS SNS

Forward Alert Manager alert to AWS SNS.

# How to use it

```bash
docker run -d --name am2sns -p 9876:9876 -e AWS_SNS_REGION=$AWS_REGION -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY -e LOG_LEVEL=info scalair/am2sns -e DRY_RUN=false
```

AWS user must have `SNS:Publish` permission.