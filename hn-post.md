I posted about clippy a couple of days ago and now I've released pasty, the companion tool that completes the bidirectional file workflow. It eliminates the cd ~/Downloads && ls -lt && cp ... && cd - workflow when you need to get a file into your project.

Together they solve both directions:

- clippy: Copy files from terminal that actually paste into GUI apps (Slack, email, etc.)

- pasty: Paste files from GUI apps into your terminal

Core use case: Copy a file in Finder → pasty → file gets copied to your current directory.

Smart text handling: Copy a text file in Finder → pasty → outputs the file's content to terminal.

Recent downloads feature:

    # Download something in your browser, then:
    pasty -r                    # copies it to current directory
    pasty -r --pick             # interactive picker for multiple files
    pasty -r 5m                 # only last 5 minutes
    pasty -r --batch            # all files from same download batch

This pasty -r feature alone has already saved me tons of time and context switching.

You get both tools: brew install neilberkman/clippy/clippy

Technical: Both tools use direct Objective-C bindings (no AppleScript). Library-first architecture, so you can use the functions in your own Go code.