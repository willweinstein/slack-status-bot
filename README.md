# slack-status-bot
Bot that autoupdates your Slack status based on your schedule and information from MyHomeworkSpace.

Still very much in beta.

## Usage
```~/work/bin/slack-status-bot --mhs-client "<MyHomeworkSpace client ID>" --slack-client-id "<Slack client ID>" --slack-client-secret "<Slack client secret>"```

You will be prompted to open your web browser if you need to set up accounts and things like that. If you need to open the management screen later, add the `--manage` flag to the command line.