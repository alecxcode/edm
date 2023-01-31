# Contributing

EDM Project uses GitHub and a development EDM System server itself (available to the core development team) to manage development process.

* If you have a simple fix or improvement, just create a pull request as described in the section below.

* If you are going to make a foundational change, first submit your ideas to the project author with [this contact form](https://edmproject.github.io/contact.html). Then wait while the core development team reviews your proposal.

* You can also become a part of the core development team or contribute with other skills than coding: e.g. design, reviews, writing articles, creating distributions. If you cannot do whatever you want without participation of the author, please, contact using the above link in relation to this matter.

## Source code contributions

1. Create a branch from the main.

2. Write your code. A commit should address a separate (encompassed by one topic or issue) modification. If possible, add tests relevant to this modification.

3. Make a pull request. Please, use one commit per request whenever possible. You can utilize `squash` in some cases to achieve that. Don't forget to make sure that your repository is up to date.

## Backend dependency management

EDM System uses [Go modules](https://go.dev/blog/using-go-modules) to manage dependencies on external packages.

If you are adding new dependencies to the packages of this project, run `go mod tidy` from the project root directory.

Then `go.mod` and `go.sum` should be added to your commit when submitting a pull request.
