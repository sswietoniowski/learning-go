# Backend Masterclass Recap

During the *Go Backend Masterclass*, we built the foundations of Three Dots Eats:
a food delivery platform with restaurants, customers, couriers, and the basic order flow.
The codebase from the end of that training is now in your workspace,
ready for the next round of features.

If you didn't take the Backend Masterclass, take a few minutes to scan the linked exercises below.
We won't cover those ideas deeply in this training because we'll focus on other things.
The AI Mentor at the bottom of the page is here to help if anything in the existing code feels unclear.

## What We Have So Far

Here's the project as it stands today, sketched out as an Event Storming board.
We use Event Storming to show how the system works.

Don't worry about the notation yet.
We'll cover it properly in the next module.

{{miroBoard "3458764665343723578"}}

Even without knowing what the colors mean,
after reading the cards you should be able to follow what the system does:
customers place orders, restaurants prepare the food, and couriers deliver it.

## Recap of What We Used

Here's a quick recap of the patterns we used during the *Go Backend Masterclass*.
If any of them looks unfamiliar, the linked exercises cover each topic in depth.

{{digression "Eight patterns from the Backend Masterclass: modular monolith, clean architecture, repositories, and more."}}

- **Modular monolith:** two modules so far, `orders` and `delivery`,
  wired into one binary by `cmd/main.go`.
  Cross-module calls are Go function calls, not network calls.
  In this training, we'll keep following the [monolith first](https://academy.threedots.tech/knowledge/monolith-first) approach we started with the
  {{exerciseLink "service scaffolding" "02-project-setup" "02-service-scaffolding" "backend-masterclass-beta"}}
  exercise in the *Go Backend Masterclass*.
  Even if we get the boundaries wrong, moving them is much easier than untangling microservices.
  And if we ever need to deploy a module separately, that's a trivial step from here.

- **Per-module Postgres schema:** each module owns its own schema in a single Postgres database.
  Migrations live in the module that owns them.
  See more in the
  {{exerciseLink "adding migrations exercise" "04-database" "01-add-migrations" "backend-masterclass-beta"}}.

- **The `common/` package:** general-purpose helpers shared across modules,
  like UUID generation, error types, HTTP plumbing, and logging.
  No application logic lives there. Anything specific to `orders` or `delivery` stays inside its own module.

- **Code generation for HTTP and SQL:** HTTP handlers are generated from an OpenAPI spec,
  and database access code is generated from SQL queries.
  It's less boilerplate to maintain by hand, and the codebase stays consistent across modules.
  See more:
  {{exerciseLink "HTTP handler from OpenAPI" "03-http" "01-http-handler" "backend-masterclass-beta"}}
  and
  {{exerciseLink "generating database code from SQL" "04-database" "02-generate-sqlc" "backend-masterclass-beta"}}.

- **Clean Architecture:** each module splits into three layers.
  `api/` holds the entry points,
  `adapters/` holds external dependencies,
  and `app/` holds business logic and orchestration.
  Dependencies flow through interfaces defined close to where they're used.
  See more:
  {{exerciseLink "application layer types" "06-application-layer" "01-app-types" "backend-masterclass-beta"}}.

- **Repositories:** they sit in the `adapters/` layer and abstract database access,
  so the `app/` layer never talks to SQL directly.
  See more:
  {{exerciseLink "implementing the repository" "05-repository" "01-implement-repository" "backend-masterclass-beta"}}
  for the basics, and
  {{exerciseLink "the closure-based update pattern" "08-advanced-repositories" "05-quotes-repository" "backend-masterclass-beta"}}
  for safely loading and modifying an entity in one transaction.

- **Component tests:** most of the test pyramid is component tests.
  We call one HTTP endpoint, then verify the result via another endpoint.
  They use Postgres and no external services.
  Unit tests cover business logic in `app/`, and integration tests cover adapters.
  See more:
  {{exerciseLink "component tests" "07-errors-and-testing" "03-component-tests" "backend-masterclass-beta"}}.

- **Read models:** when the data model we read doesn't match any existing model,
  we write a specialized SQL query and map the result straight to the HTTP response.
  We don't use the app layer because reads are just data retrieval, not business logic.
  See more:
  {{exerciseLink "the first read model" "09-read-models" "01-simple-read-model" "backend-masterclass-beta"}}
  for the basics, and
  {{exerciseLink "ordering and filtering on a read model" "09-read-models" "02-ordering-filtering" "backend-masterclass-beta"}}
  for more complex use cases.

- **Inter-module communication:** cross-module calls are Go function calls, with no network or serialization.
  Each module exports a small "contract" interface that defines what other modules can call.
  The rest is used only within the same module.
  See more:
  {{exerciseLink "calling another module" "10-inter-module-communication" "01-delivery-service" "backend-masterclass-beta"}}.

{{enddigression}}

## Environment

When you run `tdl training run`,
the platform starts everything the project needs (Postgres and the platform gateway),
builds the code, runs the tests, and reports back.
You don't need to install or configure any of that yourself.

You can also run the project locally for development.
Open the `Taskfile.yml` to see the commands we use most often:

- `task up` boots the project with Docker Compose.
- `task test` runs the full test suite (unit, integration, and component).
- `task gen` regenerates code from the OpenAPI spec and SQL queries.

## ## Exercise

Exercise path: ./project

There's nothing to change in this exercise.
The project is already in your workspace, exactly as we left it at the end of the *Go Backend Masterclass*.

If you skipped that training, take a few minutes to look around using the recap above as a map.
We encourage running the project locally too,
so you can poke at the code and run `task test` to see how the pieces fit together.

If you completed the entire *Go Backend Masterclass* including the bonus module 11,
you can replace the cloned project with your own files instead.
