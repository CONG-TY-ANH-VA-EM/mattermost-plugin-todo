# Disclaimer

**This repository is community supported and not maintained by Mattermost. Mattermost disclaims liability for integrations, including Third Party Integrations and Mattermost Integrations. Integrations may be modified or discontinued at any time.**

# Mattermost Todo Plugin

[![Build Status](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-todo/master.svg)](https://circleci.com/gh/mattermost/mattermost-plugin-todo)
[![Code Coverage](https://img.shields.io/codecov/c/github/mattermost/mattermost-plugin-todo/master.svg)](https://codecov.io/gh/mattermost/mattermost-plugin-todo)
[![Release](https://img.shields.io/github/v/release/mattermost/mattermost-plugin-todo)](https://github.com/mattermost/mattermost-plugin-todo/releases/latest)
[![HW](https://img.shields.io/github/issues/mattermost/mattermost-plugin-todo/Up%20For%20Grabs?color=dark%20green&label=Help%20Wanted)](https://github.com/mattermost/mattermost-plugin-todo/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3A%22Up+For+Grabs%22+label%3A%22Help+Wanted%22)

A plugin to track Todo issues in a list and send you daily reminders about your Todo list.

**[Help Wanted](https://github.com/mattermost/mattermost-plugin-todo/issues?utf8=%E2%9C%93&q=is%3Aopen+label%3A%22up+for+grabs%22+label%3A%22help+wanted%22+sort%3Aupdated-desc)**

## Install

1. Go the releases page and download the latest release.
2. On your Mattermost, go to System Console -> Plugin Management and upload it.
3. Start using it!

## ✨ Enterprise Features

This integration has been enhanced with enterprise-grade features for better collaboration and task management:

- **Relational SQL Backend**: Highly scalable storage supporting concurrency and complex queries.
- **Task Comments**: Engage in discussions directly on Todo items.
- **Priority Levels**: Categorize tasks by urgency (Low, Medium, High).
- **Due Dates**: Set and track deadlines with visual indicators.
- **Audit Logs**: Enterprise-ready traceability for all task modifications.
- **Internationalization**: Full support for English and Vietnamese (more coming soon).

## Usage

### Managing Todos
- **Add with Priority/Due Date**: Use the Sidebar or `/todo add` to create tasks. In the UI, you can specify a priority and a deadline.
- **Comments**: Open a task in the Sidebar to view and add comments. Manage your discussions directly within the Todo item.
- **Edit**: Modify task messages, descriptions, priorities, and due dates at any time.

### Collaboration
- **Send and Assign**: Easily delegate tasks to other team members. Receivers will get a notification via the Todo bot.
- **Status Tracking**: Transition tasks between "Open", "Accepted", and "Completed".

Every day you will get a reminder of the issues you need to complete from the `Todo` bot. The message is only sent if you have issues on your Todo list.

## Development

This plugin contains both a server and web app portion. Read our documentation about the [Developer Workflow](https://developers.mattermost.com/integrate/plugins/developer-workflow/) and [Developer Setup](https://developers.mattermost.com/integrate/plugins/developer-setup/) for more information about developing and extending plugins.

### Releasing new versions

The version of a plugin is determined at compile time, automatically populating a `version` field in the [plugin manifest](plugin.json):
* If the current commit matches a tag, the version will match after stripping any leading `v`, e.g. `1.3.1`.
* Otherwise, the version will combine the nearest tag with `git rev-parse --short HEAD`, e.g. `1.3.1+d06e53e1`.
* If there is no version tag, an empty version will be combined with the short hash, e.g. `0.0.0+76081421`.

To disable this behaviour, manually populate and maintain the `version` field.
