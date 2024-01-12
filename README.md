# Gowind CSS a *simple* utility-first CSS framework for rapidly building custom user interfaces

# Documentation
This project attempts to be very compatible with Tailwind CSS.
Check out their documentation https://tailwindcss.com/docs/utility-first
This project is incomplete but I expect to quickly add support for everything other than the following things:
similarities with tailwind
- all the same default base classes
- almost all the same default variants (except for *: and []:)
- arbitrary values for base classes
- arbitrary values for variants
- theme() and spacing() functions
- group and peer support

# Why gowind instead of tailwind?
I created gowind because I wanted to use tailwind css in a project but to get autocomplete support in my Jetbrains IDE I had to install tailwind through npm.
Previously I was using their standalone cli.
This really irked me because I had to add another language, have a giant node_modules folder, have a tailwind.config.js, and update my .gitignore just for some simple autocomplete.
After that I began experimenting with creating my own version of tailwind.
During that process I did some more research into tailwind and watched [the part of the tailwind conference](https://youtu.be/CLkxRnRQtDE?t=2146&feature=shared) where they announce that they will be switching to Rust for their engine.
I think Rust is a great language and a great choice for a project like this.
However, since tailwind's configuration file and plugin system is in JavaScript, they will never be able to drop that language or Node.js and now they will be stuck with two langauges for the codebase.
This might be the right choice in the long run, but it seems to me to be adding too much complexity.
I am unsure about tailwind's direction so that is why I am interested in building gowind.
It is a much simpler alternative that handles 90% of what most people use tailwind for at less than 10% of the complexity.
The best part is that you can always just go back to writing CSS lol, this tool does not need to be that powerful.

Compared to tailwind, gowind has (or will have):
- a built-in html formatter (maybe add linter)
- built-in ide support (dump base classes)
- a built-in playground for testing your concoctions
- built-in documentation (offline)
- no @apply to shoot yourself in the foot with
- an arbitrary base class so you can just write inline css without writing inline css
- It is built with only dependency on golang standard library for the distributed binary
- built with only one language (tailwind is built with not only JavaScript but also Rust, LOL)
- much faster
- much smaller binary size
- much, much simpler
- the entire source code is included with the distributed binary, so you can easily hack on it if you have a go compiler

# Why tailwind instead of gowind
- developed by multiple professionals instead of a random guy who had free time during winter break
- gowind does not support *: or []: (both could be doable but seems like a lot of work that I will put off until someone submits a pr and gives me at least 100$ to do it.)
- gowind does not have external plugins

# Contributing
If you are interested in contributing just submit a PR and explain what and why you made those changes.
Try to follow the style of the existing code, however, if your style is better, then explain why and I might switch to it.
Please add test cases for whatever you changed though, the syntax is pretty set in stone so we really just want a lot of tests.

# Design Decisions and Developer Notes
The entire codebase is in two files.
This is a bit unorthodox so let me try to convince myself why it is a good idea.
1. It is different and I wanted to try something new.
2. It is not so bad if you fold everything.
3. It encourages the total number of functions and structs to remain small as that is the main thing that grows the size of the file after folding.
4. Less files means less places to look.
5. Puts more pressure on keeping everything small.
6. I should not need more than one file considering how simple this tool _should_ be

I am not sure if the / syntax should exist.
It just makes parsing harder and creates one more thing for the user to think about.
I think it would be better if we did multiple arbitrary variants.
I am going to implement it anyways just to say that I really supported all the features.
However, if I will turn this into a real tool, I might remove some cruft of the language before 1.0.

I think tailwind css uses string concatenation to build up their selectors.
I use fields on a struct for different parts of the selector then combine them and turn it into a string when I want to print it.
The difference between these two methods prevents me from doing the "[]:" arbitrary variant as easily as they can.
I will skip over this feature for now but it could always be introduced as just another variant plugin.
Although I would have to create a special parser to put the arbitrary data into my struct.
That is too much complexity, and I am not sure how useful this feature is.
We already support arbitrary variants, so if we had a great selection of base variants then "[]:" would not be needed.
The same could be said about the arbitrary declaration base class, but I am keeping it because of how simple it was to implement.

Tailwind has some crazy things that are crazy to implement.
Namely *: because it shifts the target of the variants to a > * instead of the first selector.
Also []: because it does a lot of crazy stuff with the string.
I am just going to have to accept defeat and do an 80/20 solution.
When you are doing something that complex anyways it would probably be good to just use normal css.
I could probably support all of tailwind but it would be too crazy.

What is the essence of tailwind?
- utility classes
- locality of behavior
- no need to spend time naming
- best of both convention and customization