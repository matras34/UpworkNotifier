// ==UserScript==
// @name         Upwork Keyword Notifier with Tags Check
// @namespace    http://tampermonkey.net/
// @version      1.9
// @description  Keyword search on Upwork with title, description, and tags check, sends notifications to Telegram, and saves data between page reloads using localStorage.
// @author       ChatGPT
// @match        *://www.upwork.com/*
// @grant        none
// ==/UserScript==

(function() {
    'use strict';

    const keywords = ["keyword1", "keyword2", "keyword3"]; // replace with your keywords
    const telegramBotToken = "YOUR_TELEGRAM_BOT_TOKEN";
    const telegramChatId = "YOUR_TELEGRAM_CHAT_ID";
    
    // Load the list of sent job IDs from localStorage, or create a new one if it's empty
    const sentJobs = JSON.parse(localStorage.getItem('sentJobs') || '[]');

    // Function to save the list of sent job IDs to localStorage
    function saveSentJobs() {
        localStorage.setItem('sentJobs', JSON.stringify(sentJobs));
    }

    // Function to send notifications to Telegram
    function sendTelegramNotification(message) {
        const url = `https://api.telegram.org/bot${telegramBotToken}/sendMessage?chat_id=${telegramChatId}&text=${encodeURIComponent(message)}`;
        
        console.log("Sending request to Telegram:", url);

        fetch(url)
            .then(response => response.json())
            .then(data => {
                if (data.ok) {
                    console.log("Notification successfully sent to Telegram");
                } else {
                    console.error("Error sending notification:", data);
                }
            })
            .catch(error => {
                console.error("Network error or request issue:", error);
            });
    }

    // Function to check a job for keywords
    function checkJobForKeywords(jobElement) {
        const jobId = jobElement.getAttribute("data-ev-job-uid");

        // Check if a notification has already been sent for this job
        if (sentJobs.includes(jobId)) {
            console.log(`Job with ID ${jobId} has already been notified, skipping.`);
            return;
        }

        const titleElement = jobElement.querySelector('a[data-test*="job-tile-title-link"]');
        const descriptionElement = jobElement.querySelector('div[data-test="UpCLineClamp JobDescription"] p');
        const tagsElements = jobElement.querySelectorAll('.air3-token-container .air3-token-wrap span');

        if (!titleElement) {
            console.warn("Job title not found in element:", jobElement);
            return;
        }
        
        if (!descriptionElement) {
            console.warn("Job description not found in element:", jobElement);
            return;
        }

        const titleText = titleElement.textContent.toLowerCase();
        const descriptionText = descriptionElement.textContent.toLowerCase();

        // Combine text from all tags into a single string for searching
        const tagsText = Array.from(tagsElements).map(tag => tag.textContent.toLowerCase()).join(" ");

        console.log("Checking job with title:", titleText);

        let foundKeywords = keywords.filter(keyword => 
            titleText.includes(keyword.toLowerCase()) || 
            descriptionText.includes(keyword.toLowerCase()) || 
            tagsText.includes(keyword.toLowerCase())
        );

        if (foundKeywords.length > 0) {
            console.log(`Keywords found: ${foundKeywords.join(", ")} in job "${titleText}"`);
            
            const jobLink = titleElement.href;
            const jobTitle = titleElement.textContent;
            const message = `Job with keywords found: ${foundKeywords.join(", ")}\nTitle: ${jobTitle}\n\nLink: ${jobLink}`;
            
            console.log("Message for Telegram:", message);

            // Send notification, add ID to the list of sent jobs, and save it
            sendTelegramNotification(message);
            sentJobs.push(jobId);
            saveSentJobs();
        } else {
            console.log("No keywords found in this job");
        }
    }

    // Main function to scan all jobs on the page
    function scanJobs() {
        const jobContainer = document.querySelector('.card-list-container');
        if (!jobContainer) {
            console.error("Job container not found");
            return;
        }

        const jobs = jobContainer.querySelectorAll('article.job-tile');
        console.log(`Found ${jobs.length} jobs to check`);
        
        jobs.forEach(job => checkJobForKeywords(job));

        // After checking all jobs, wait 3 seconds and reload the page
        setTimeout(() => {
            console.log("Reloading page to get new jobs...");
            location.reload();
        }, 3000); // 3-second wait before reloading
    }

    // Wait for content to load, then start scanning
    setTimeout(() => {
        scanJobs();
    }, 5000); // Adjust waiting time depending on page load speed

})();
