<div align="center">
<h1>DI - Developer Interface</h1>

</div>
<br/>

## Table of Contents <!-- omit in toc -->

- [Overview](#overview)
- [Usage](#usage)
  - [devint](#devint)
  - [devint git createpr](#devint-git-createpr)
  - [devint git summarizepr](#devint-git-summarizepr)
  - [devint config](#devint-config)
- [Configuration](#configuration)


## Overview

The Developer Interface (DI) is a command-line tool designed to streamline
developer workflows. DI helps developers quickly perform routine operations and
maintain consistency across projects.

## Usage

The Developer Interface (DI) enables streamlined development workflows by providing a unified command-line interface **to** manage configuration settings, execute Git operations, and more. Below are tables of available commands and their flags:

### devint

| Flag         | Type | Required | Description                          |
| ------------ | ---- | -------- | ------------------------------------ |
| -t, --toggle | bool | ❌        | Toggle verbose mode or other options |
| -h, --help   | bool | ❌        | Show help for devint                 |

### devint git createpr

| Flag                     | Type   | Required | Description                                                                         |
| ------------------------ | ------ | -------- | ----------------------------------------------------------------------------------- |
| --pr-title (-t)          | string | ✅        | PR title. Will open a draft PR if the string contains [DRAFT] or [WIP]              |
| --target-branch (-b)     | string | ❌        | Target branch (default "main")                                                      |
| --issue (-i)             | int    | ❌        | Issue number                                                                        |
| --dummy (-d)             | bool   | ❌        | Dummy mode. Will print summary to console and clipboard but not open a PR on GitHub |
| --provider-override (-p) | string | ❌        | LLM provider override. Sets the LLM provider only for this request                  |
| --model-override (-m)    | string | ❌        | LLM model override. Sets the LLM model only for this request                        |

### devint git summarizepr

| Flag              | Type   | Required | Description                                                                    |
| ----------------- | ------ | -------- | ------------------------------------------------------------------------------ |
| --pr-number (-p)  | int    | ✅        | Pull request number to summarize                                               |
| --repo-owner (-r) | string | ❌        | GitHub repository owner override. If set, overrides the repo_owner from config |

This command fetches a pull request by number, retrieves its diff, and saves a formatted markdown summary to the directory specified by `pr_summary_output_dir` in your config file. The summary includes PR metadata (title, number, status, author, URL), description, and formatted diff.

**Note:** The `pr_summary_output_dir` must be set in your `git_config` (via `devint config`) for this command to work.

### devint config

| Flag          | Type   | Required | Description                                                                  |
| ------------- | ------ | -------- | ---------------------------------------------------------------------------- |
| --show (-s)   | bool   | ❌        | Show the configuration                                                       |
| --editor (-e) | string | ❌        | Edit the configuration in the given text editor, for example `nano` or `vim` |

To run the interactive editor, run without any flags, ie. `devint config`.

## Configuration

The configuration is done through a config YAML file located at `~/.devint.config.yaml`.

An example configuration file can be found at `config/examples/.config.example.yaml`.

Configuration can be updated using the interactive command `devint config`.

