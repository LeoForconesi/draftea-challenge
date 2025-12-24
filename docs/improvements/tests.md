# Integration Tests
We want to add integration tests to verify the end-to-end functionality of the Wallet API service, 
including interactions with the database and external services. Below is a suggested structure and example for implementing integration tests in Go.

## Test Structure
- Use a separate test database to avoid interfering with production data.
- Use Docker Compose to spin up a test environment with PostgreSQL and RabbitMQ.
- Create test data builders to simplify the creation of test entities.
- Use the `testing` package along with `httptest` to test HTTP endpoints.
- We already use a mocking gateway for payments, which can be utilized in integration tests.
- Provide a sepparate configuration file for integration tests.
- Include these results in a separate workflow in CI/CD pipeline, configuring as a check to be passed before merging code to main branch.
- I'm inclined to use a different repository for integration tests to keep them isolated from unit tests and the main application codebase. 
This separation helps maintain a clear boundary between the application logic and the tests that validate its behavior in a real-world scenario. 
It also allows for independent versioning and deployment of the test suite, which can be beneficial for larger projects or teams, 
and have for disposition a dedicated CI/CD pipeline and specific libraries for testing, avoiding any conflict with the main application dependencies.
- Among the functionalities we can include by configuration files and test suites, some stress tests to validate the performance of the application under load. 
This only are triggered manually.
- If we use a different repository we could even use different languages or frameworks for testing, if needed. Python has proven to be a great choice for 
testing due to its simplicity and the rich ecosystem of testing libraries available.


## Other mentions
Always is useful to integrate slack messages with the CI/CD pipeline to notify the team about the results of the integration tests, especially if they fail.
This way we can ensure that any issues are promptly addressed before merging code changes

Any integration test or stress test should be executed in a stage environment, or locally, never in production for obvious reasons.

