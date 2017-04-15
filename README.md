# slack-status-bot
Bot that autoupdates your Slack status based on your schedule and information from MyHomeworkSpace.

Still very much in beta.

Currently only supports one user per process/working directory, and would most likely require some refactoring to work with multiple users at the same time.

## Usage
### Command line
```./slack-status-bot --mhs-client "<MyHomeworkSpace client ID>" --slack-client-id "<Slack client ID>" --slack-client-secret "<Slack client secret>"```

You will be prompted to open your web browser if you need to set up accounts and things like that. If you need to open the management UI later, add the `--manage` flag to the command line. Note that the bot will not start until you click the link in the management UI.

The bot should be left running in the background. You should not run it on a laptop or something that could fall asleep/hibernate/etc, as that could mess with the timing and cause statuses to not be set correctly.