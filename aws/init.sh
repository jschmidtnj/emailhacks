#!/bin/sh

REGION=us-east-1

aws iam create-role --region $REGION --role-name mailpear-task-role --assume-role-policy-document file://mailpear-ecs-task-role-trust-policy.json
aws iam create-role --region $REGION --role-name mailpear-task-execution-role --assume-role-policy-document file://mailpear-ecs-task-role-trust-policy.json
aws iam put-role-policy --region $REGION --role-name mailpear-task-role --policy-name mailpear-iam-policy-task-role --policy-document file://mailpear-iam-policy-task-role.json
aws iam put-role-policy --region $REGION --role-name mailpear-task-execution-role --policy-name mailpear-iam-policy-task-execution-role --policy-document file://mailpear-iam-policy-task-execution-role.json
aws ecs register-task-definition --region $REGION --cli-input-json file://mailpear-task.json
