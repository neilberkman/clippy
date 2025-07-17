I posted about clippy a couple of days ago and now I've released pasty, the companion tool that completes the bidirectional file workflow.

Together they solve both directions:

- clippy: Copy files from terminal that actually paste into GUI apps (Slack, email, etc.)

- pasty: Paste files from GUI apps that actually work in your terminal

Core use case: Copy a file in Finder → pasty → file appears in your current directory.

Smart text handling: Copy a text file in Finder → pasty → outputs the file's content to terminal.

Since my last post, I've also added some killer features to clippy:

Recent downloads with interactive picker:

    clippy -r                   # copy most recent download
    clippy -r --pick            # interactive picker
    clippy -r --paste           # copy and paste in one step

You get both tools: brew install neilberkman/clippy/clippy

Technical: Both tools use direct Objective-C bindings (no AppleScript). Library-first architecture, so you can use the functions in your own Go code.