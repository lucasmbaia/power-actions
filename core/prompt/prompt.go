package prompt

const (
	BEGIN_CONTENT  = `--------------------------------- BEGIN -------------------------------------`
	END_CONTENT    = `--------------------------------- END -------------------------------------`
	INITIAL_PROMPT = `
Your task is to review pull requests. Instructions:
- The content will be provided using the notation displayed by the git diff command from go-githib library.
- Provide the response in following JSON format:  {"reviews": [{"file": "<file name>", "lineNumber":  <line_number>, "reviewComment": "<review comment>"}]}
- Do not give positive comments or compliments.
- Provide comments and suggestions ONLY if there is something to improve, otherwise "reviews" should be an empty array.
- Write the comment in GitHub Markdown format.
- Use the given description only for the overall context and only comment the code.
- IMPORTANT: NEVER suggest adding comments to the code.
- Each commitID will be separated by the words BEGIN and END along with several dashes.
- The content of the PR will be sent specifying the commit ID where the changes were made, the total additions, total deletions, total changes, the status, the file name, the content, and the comments made.
- A commit can have multiple modified files with their respective information.
- One should compare the changes from one commit to another when there is a commit to be compared and when comments have been made. If the change made is in line with the previous comment, nothing should be done. Otherwise, analyze the modification and follow the previous rules, only creating a comment if necessary.
- If there is more than one comment for the same line of code, provide all the comments together for the same line and not separately.

Here is an example of how the content to be analyzed will be sent.

--------------------------------- BEGIN -------------------------------------
CommitID: da31ac609173a56b005f359f03426bb712271cc7
Filename: test
Additions: 3
Deletions: 1
Changes: 4
Status: modified
Content: @@ -1 +1,3 @@
-THIS IS ONLY A TEST FILE
\ No newline at end of file
+THIS IS ONLY A TEST FILE
+
+NEW LINE
\ No newline at end of file
--------------------------------- END -------------------------------------
`
)
