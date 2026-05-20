# 🚀 Git & GitHub Guide

This document explains the core Git operations, how to use them in VS Code, and the collaborative workflow used in this repository.

---

## 1. Core Git Concepts

### `git clone`
- **What it does:** Creates a local copy of a remote repository on your machine.
- **Why it's useful:** It's the first step to start working on a project.
- **Command:** `git clone <repository-url>`

### `git fetch`
- **What it does:** Downloads the latest history and metadata from the remote repository but **does not** change your local code.
- **Why it's useful:** Allows you to see what others have done without risking conflicts in your current work.
- **Command:** `git fetch origin`

### `git pull`
- **What it does:** Downloads changes from the remote repository and immediately tries to merge them into your current branch. (`pull` = `fetch` + `merge`).
- **Why it's useful:** Keeps your local branch up to date with the team's progress.
- **Command:** `git pull origin <branch-name>`

### `git commit`
- **What it does:** Records your changes in the local repository's history with a descriptive message.
- **Why it's useful:** Creates a "save point" you can return to if things go wrong.
- **Command:** `git add .` (stages changes) then `git commit -m "feat: descriptive message"`

### `git push`
- **What it does:** Uploads your local commits to the remote repository (GitHub).
- **Why it's useful:** Shares your work with the rest of the team.
- **Command:** `git push origin <branch-name>`

### `git stash`
- **What it does:** Temporarily "hides" your uncommitted changes to give you a clean working directory.
- **Why it's useful:** Useful when you need to switch branches quickly but aren't ready to commit your current work.
- **Commands:** 
  - `git stash` (save changes)
  - `git stash pop` (restore and remove from stash)
  - `git stash list` (see all stashed work)

---

## 2. Git in VS Code

VS Code provides a built-in GUI for Git that makes these operations easier to visualize.

1. **Source Control View:** Click the icon on the left sidebar (shortcut: `Ctrl+Shift+G`).
2. **Staging:** Click the `+` icon next to a file to stage it (equivalent to `git add`).
3. **Committing:** Type your message in the text box at the top and click **Commit**.
4. **Pushing/Pulling:** Click the "Sync" icon in the bottom-left status bar or use the `...` menu in the Source Control view.
5. **Stashing:** Open the Command Palette (`Ctrl+Shift+P`), type "Git Stash", and select the desired action.

---

## 3. GitHub Workflow: Pull Requests & CI

In professional environments and this project, we follow a strict workflow to maintain code quality.

### 🛑 Branch Protection
**Never commit directly to the `main` branch.** It is protected. Any attempt to `git push origin main` will likely be rejected by GitHub.

### 🛠️ The Feature Branch Workflow
1. **Create a branch:** `git checkout -b feature/your-feature-name`.
2. **Work and Commit:** Make your changes locally.
3. **Push:** `git push origin feature/your-feature-name`.
4. **Open a Pull Request (PR):** Go to the repository on GitHub. You will see a button to "Compare & pull request".

### 🧪 Pull Requests & CI (Continuous Integration)
When you open a PR, our **CI Pipeline** (GitHub Actions) automatically starts:
- It runs tests, linters, and build checks.
- **The "Green Light":** You will see a status section at the bottom of the PR page. 
- **Blocking Merges:** If any check fails (indicated by a red ❌), the **Merge** button will be disabled. You must fix the errors and push again until you see a green checkmark ✅.

### 🤝 Merging
Once the CI is "green" and you have received a peer review:
1. Click **Merge Pull Request**.
2. Select **Squash and Merge** (preferred) to keep the `main` history clean.
3. Delete your feature branch.
