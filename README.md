# Upwork Keyword Notifier with Tags Check

## Description
This Tampermonkey script automates job search on [Upwork](https://www.upwork.com) based on specified keywords. It scans the job title, description, and tags for matching keywords and then sends notifications to Telegram for found matches. The script uses `localStorage` to keep track of sent notifications and reloads the page after each scan to continue monitoring for new jobs.

## Features
- Scans Upwork job titles, descriptions, and tags for specified keywords.
- Sends notifications to a specified Telegram chat when keywords are found.
- Uses `localStorage` to store data on jobs that have already been notified, preventing duplicate notifications.
- Automatically reloads the page 3 seconds after finishing each scan to continue monitoring for new jobs.

## Installation
1. Install the [Tampermonkey](https://www.tampermonkey.net/) extension for your browser.
2. Create a new script in Tampermonkey and paste in the code from `Upwork Keyword Notifier with Tags Check`.
3. Replace the following values in the script:
   - `keywords` — replace with your desired list of keywords.
   - `telegramBotToken` — replace with your Telegram bot's API token.
   - `telegramChatId` — replace with the ID of the Telegram chat where notifications will be sent.
4. Save the script and open Upwork. The script will start working automatically.

## Configuration
- **Keywords**: Update the `keywords` array with the words you want to search for.
- **Telegram Token and Chat ID**: Replace `telegramBotToken` and `telegramChatId` with your own values.
- **Page Reload Delay**: By default, the script reloads the page 3 seconds after scanning. You can adjust the `3000` value in the code to increase or decrease the wait time.
  
## How it Works
1. The script locates jobs on Upwork within the `.card-list-container` section.
2. For each job, it checks:
   - Title (`a[data-test*="job-tile-title-link"]` selector).
   - Description (`div[data-test="UpCLineClamp JobDescription"] p` selector).
   - Tags (`.air3-token-container .air3-token-wrap span` selector).
3. If any of these elements contain a keyword, the script sends a notification to Telegram.
4. Jobs that have already been notified are stored in `localStorage` and are not processed again.
5. Once all jobs have been checked, the script reloads the page 3 seconds later to update the list.

## Dependencies
- **Telegram Bot API**: The script uses this API to send notifications to Telegram.
- **LocalStorage**: Used to store data on sent notifications.

## Limitations
- The script only works in the browser and requires the Upwork page to be open.
- To run continuously, the browser and computer must be kept on.
  
## Support
If you have any questions or issues with the script, please open an issue in this repository or contact us.
