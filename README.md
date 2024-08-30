## How to Install
1. Clone Repo
2. Update 'bin/file_link.ts' file to associate this project with your AWS account or Create a Profile (Prerequisite: making it possible to access your AWS account via (AWS)CLI)
3. ```npx cdk bootstrap```
4. ```npx cdk deploy (--your_stack_name) (--your_profile)```

## Architecture
<img width="500" alt="スクリーンショット 2024-08-30 12 34 54" src="https://github.com/user-attachments/assets/99d44e34-00f4-47bc-b147-d056a2828895">

*This repo is the backend part of the project. The frontend is right here: https://github.com/shokishimo/FileLink-Web
