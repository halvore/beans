# /ship — Prepare and Open a Pull Request

You are preparing code for a pull request. Your job is to make sure everything is ready: tests pass, beans are updated, and the PR is well-described.

## Process

1. **Check the state.** Run `git status` and `git diff` to understand what's changed. Identify which bean(s) this work relates to.

2. **Run tests.** Execute `mise test` to run the full test suite. If any tests fail, fix them before proceeding.

3. **Check for frontend warnings.** If frontend files were changed, run `pnpm build` from the `frontend/` directory and resolve any compiler warnings.

4. **Update beans.** For each bean involved:
   - Check off completed todo items
   - Add a `## Summary of Changes` section if the bean is being completed
   - Move to `review` status if the work is done, or keep `in-progress` if partial

5. **Commit any remaining changes.** Make sure all changes (including bean updates) are committed with proper conventional commit messages that reference bean IDs.

6. **Open the PR.** Create a pull request using `gh pr create` with:
   - A clear title following conventional commit format, including bean ID(s)
   - A summary section with bullet points describing the changes
   - A test plan section describing how to verify the changes
   - Link to relevant bean ID(s)

## Rules

- Never ship with failing tests
- Never skip the test run — even if "nothing changed"
- Include bean file changes in the commits
- Follow the project's commit message conventions (conventional commits with bean refs)
- The PR description should help a reviewer understand the "why", not just the "what"
