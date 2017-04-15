# slack-status-bot
Bot that autoupdates your Slack status based on your schedule and information from MyHomeworkSpace.

Still very much in beta.

Currently only supports one user per process/working directory, and would most likely require some refactoring to work with multiple users at the same time.

## Usage
### Command line
```./slack-status-bot --mhs-client "<MyHomeworkSpace client ID>" --slack-client-id "<Slack client ID>" --slack-client-secret "<Slack client secret>"```

You will be prompted to open your web browser if you need to set up accounts and things like that. If you need to open the management UI later, add the `--manage` flag to the command line. Note that the bot will not start until you click the link in the management UI.

The bot should be left running in the background. You should not run it on a laptop or something that could fall asleep/hibernate/etc, as that could mess with the timing and cause statuses to not be set correctly.

### MyHomeworkSpace
Every day at midnight (and when you first run it), the bot will scan for events in MyHomeworkSpace. It looks for a class called "Other", and then in that class, for homework items with the prefix "BuildSession". It then reads the description of these items. They should be formatted like this (case and space sensitive!):

```
Name: Some build session
Room: 501
Start: 3:15pm
End: 5:00pm
```

While the "Start" and "End" parameters are required, the "Name" and "Room" parameters are optional, and will default to "Build session" and "501" respectively. If there is an error parsing the decription, it will be logged to stdout.