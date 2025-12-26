# Use claude-crc, Don't Lose Features to 'Generous' Inference
## *3-Level Spec-Driven Development For Claude*

## The Problem With 2-Level Spec-Driven Development

2-level spec driven development suffers from the same problem as vibe coding -- it just takes longer. Which means it bites you after you done a lot more work. I did some hand-rolled, 2-level spec-driven development for an app I've been building. By 2-level, I mean you have a collection of spec documents and tell Claude to generate code from it. When you change one set, you tell Claude to check your changes and update the other. It works pretty well as a light-weight alternative to vibe coding. But it's not a tiny app and I after a while I ran into problems.

Claude "generously" inferred things from my specs that weren't actually in there. I wasn't as diligent as I should have been with reviewing the generated code -- because AI, I suppose. It generates so much code so fast that eventually it becomes easier just to have it auto-approve already. Claude can give you a lot more than you asked for. It's great! But then I changed a spec, Claude changed code in response to that, and it decided to remove one of the generous additions I hadn't asked for. Only I had come to think of that as part of my app.

## The Solution: 3-Level Development

What I wanted was a middle level where I could see how Claude was interpreting my specs but that wasn't reams and reams of code. I looked at some spec-driven development packages and read  [What You See Is What It Does paper](https://arxiv.org/abs/2508.14511), by Meng and Jackson at MIT. It's a good paper with clear benefits but I didn't like their "dumb data" / "procedural database code" model. Eventually, I realized that I had actually used a system kind of like theirs for decades but for human development, not Claude development.

From the late 80s through the early 2000s, I used [CRC cards](https://en.wikipedia.org/wiki/Class-responsibility-collaboration_card), sequence diagrams, and UI mockups to communicate with customers in order to verify that we were on the same page and actually understood what they really wanted. Customers are notoriously bad at telling developers what they want.

I realized with Claude as the developer I am now the customer. 😮

OK, maybe it's not the same thing. As the lead developer on this project I want to know what Claude is going to do and why. Of course this is a bit of a fantasy given how LLMs work, but they're great at following plans. This is a way to ask Claude to make a plan that's easy to examine and that makes a ton of sense for developing software.

### Before And After

I developed a game app with a 2-level process. The game is quite elaborate, it combines multiuser role playing with dynamic world building and peer-to-peer networking. Claude put a back button on one of my pages and after some spec changes, the button disappeared. I hadn't requested the button but I had come to rely on it during my testing.

The 3-level process eliminates problems like these because Claude makes **easy to read** designs reflecting my specs. This lets me easily see where the buttons are without having to run the app or pore through reams of HTML templates. And I can rest assured that the buttons in the designs will come through in the implementation. The automatic gap analysis points out where elements of the design differ from the specs and even lays out potential oversights like global error handling, security, and so on.

## Humans Doing CRC Spec-Driven Development

CRC cards, sequence diagrams, and UI mockups are great ways to talk to customers because they form a super high level design document that captures the spec and determines the implementation but despite that power, they're concise, they're visual, and they're not algorithmic. That means they're also great for communicating between customers and developers, i.e. me and Claude. They appear to be OO (object oriented) but they're actually not. They're really about ADTs (abstract data types); a concept that predates OO and permeates programming of all sorts -- basically it's just a collection of state and a set of operations on it. So they can really apply to any project, not just object oriented ones.

## The CRC Design Agent For Claude: Providing A Lightweght, Complete Software Process

The idea is to make a lightweight but complete software process for Claude that gives people a way to see how the LLM interprets the requirements so they can clarify their specifications when the LLM infers things that shouldn't be there or interprets things in ways in which you do not expect. There's another motivation for this process as well: maintainability. It'd be nice if you could be confident that when you make changes to your requirements, the changes will ripple safely down to the code.

Requirements, and design specification documents are key parts of the software process but projects also have code, tests, and documentation. This is all linked back to the specs the human writes (the "requirements") through the specs Claude writes (CRC cards, sequence diagrams, UI layouts and information). This linkage supports "traceability" so that if you change something, you can also update the things it's linked to. Claude is great for this, of course, so you can tell it, "I just changed `specs/person.md`, upate the associated design specs and code.

CRC cards, sequence diagrams, and UI layouts (both in ASCII art) are generally written by Claude but people can also write and/or edit them and then say, "I just changed `design/crc-Person.md`, upate the associated code and specs".

So traceability allows Claude to propagate changes bi-directionally. It also makes it easier and safer to refactor code, requirements, and systems.

The process also provides valuable gap analysis to show you how your specs, the design specs, and the implementation diverge and what you might have left out from your specs.

## What It Is

It's a collection of skills, commands, and agents for Claude that assist you in spec-driven development with a CRC process. You can load it into a fresh project or you can "patch it" into an existing project and ask Claude to reverse-engineer design and spec docs.

It will make entries in your project's `.claude/commands`, `.claude/skills`, `.claude/scripts`, `.claude/agents`, and `.claude/docs` and it will add some things to your project's CLAUDE.md file. It relies on two top-level directories: `specs` and `design`.

## Reverse Engineering

You can use this toolkit to reverse engineer design and specs for an existing project -- the documentation in the .claude directory provides enough context for that. Try telling Claude, "check .claude/docs and agents for context and reverse-engineer a design directory for this project".

You can try this out with the example contacts app in the project's `expected-result` directory by copying it into a fresh project and removing the `specs`, `design`, and `docs` directories. Then initialize the crc tooling and tell claude to make them itself.

## Actually Using It

There are three steps to this process and the result will be a fully linked set of levels:

- Level 1: general specification / requirements
- Level 2: the design and test designs
- Level 3: code and docs

The files in layers 2 and 3 will contain comments linking them back to the layers above them and the design directory will contain traceability files that link layers 1 and 2 to the layers below them.

### Installing It In Your Claude Project

1. Change directory into your Claude project (i.e. a directory that contains a .claude subdirectory).
2. If you don't have it, grab the [latest release](https://github.com/zot/claude-crc/releases) if needed.
3. Run the init script

```bash
# Download
curl -O https://github.com/zot/claude-crc/releases/latest/download/claude-crc-dist.py

# Run
python3 claude-crc-dist.py
```

This initializes your project. Don't worry about overwriting things -- it will refuse to run if you have conflicting files in your .claude directory.

Documentation: see the [project readme](https://github.com/zot/claude-crc).

### Step 1: General Specification / Requirements

A great way to start out is to ask Claude for help like, "I want to make a contact management app. Put specs for it into the top-level specs directory that I can use with the designer agent." Claude will generate specs for you. Take look at those and if anything is missing like, maybe, storage, you can add it yourself or you can ask Claude to add that to the specs.

### Step 2: the Design and Test Designs

When you're pretty happy with the specs ask claude, "generate designs and test designs". Claude will delegate that to its `designer` subagent and you can examine the designs Claude comes up with -- the UI layouts, sequence diagrams, CRC cards, and test designs. From examination you might find that your specs need more fleshing out (you can ask Claude to do that, of course). You might find that Claude interpreted your specs in unexpected ways -- ask Claude to explain, then correct it and tell it to update your specs so it doesn't happen again.

### Step 3: Code and Docs

When you're pretty happy with the design, ask claude, "generate code, tests, and docs" and it will populate the source code directories and the docs directory.

### Testing and Fixing Your App

Your app might (will probably) not work as soon as Claude generates it. There might (will probably) be a few things here and there that use the wrong directory, don't import something they need, and so on. Even after the tests run. You'll need to try it out and go through a test/fix loop. You'll probably benefit from using the Serena MCP and, if your UI is a web app, the Playwright MCP -- using these is well documented on the web.

## The Result: Transformative Claude Development

Maybe you've been vibe coding, 2-level spec coding, or using a spec-driven process that doesn't provide an easy-to-read middle layer, traceability, gap analysis, or doesn't support continued development very well. If so, give claude-crc a try and see how it works.  You'll get a project that you can change along the way without gaining weird features you didn't want and without losing ones you found out you wanted but didn't actually ask for.

-  **Readable designs** CRC cards show Claude's interpretation before code
-  **Traceability** Every file links back to requirements
-  **Gap analysis** Automatic detection of spec/design/code drift
-  **Bidirectional updates** Change specs, design, or code and propagate

<!-- end of list -->

-  **Get Started:** [Download v1.0.0](https://github.com/zot/claude-crc/releases/latest)
-  **Documentation:** [GitHub](https://github.com/zot/claude-crc)
-  **Questions?** [Open an Issue](https://github.com/zot/claude-crc/issues)

<!-- end of list -->


