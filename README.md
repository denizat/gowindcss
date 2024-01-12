In case .md file extension is unfamiliar to you check out this nice overview: https://www.markdownguide.org/getting-started/.
This will become a readme with documentation, just not yet.

I changed the codebase to be in a single go file.
This is quite unorthodox so let me try to convince myself why it is a good idea.
1. It is different and I wanted to try something new.
2. It is not so bad if you fold everything.
3. It encourages the total number of functions and structs to remain small as that is the main thing that grows the size of the file after folding.
4. Less files means less places to look.
5. Puts more pressure on keeping everything small.
6. I should not need more than one file considering how simple this tool _should_ be


similarities with tailwind
- all same default base classes
- almost all same default variants
- ordered classes
- arbitrary values for base classes
- arbitrary values for variants
- marker variant support
- theme() and spacing() functions
- group and peer support

differences from tailwind

pro
- built in html formatter (maybe add linter)
- built in ide support (dump base classes)
- built in playground
- built in documentation (offline)
- no @apply
- [] <- arbitrary base class (could be horrible too though)

- built with only dependency on golang standard library for the distributed binary
- built with only one language
- much faster
- much smaller binary size
- much simpler

con
- not as well developed
- no support for *: or []: (both could be doable but seems like a lot of work that I will put off until someone submits a pr and gives me at least 100$ to do it.)
- no @apply
- no external plugins

neuter
- tailwind.config.js -> gowind.json


I am not sure if the / syntax should exist.
It just makes parsing harder and creates one more thing for the user to think about.
I think it would be better if we did multiple arbitrary variants.
I am going to implement it anyways just to say that I really supported all the features.
However, if I will turn this into a real tool, I might remove some cruft of the language before 1.0.

I think tailwind css uses string concatenation to build up their selectors.
I use a selector struct and turn it into a string when I want to print it.
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
