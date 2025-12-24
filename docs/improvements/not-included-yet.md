# Not Included Features
↩️ [Return to README](../../README.md)

- I could have used Wiki documentation from github, but I preferred markdown files for easier version control, offline access, and I wrote them in while working on the project. This can easily be changed later if needed.
- I've tried to be concise and thorough on how the git repository was handled, but in order to save time it didn't result as I've would have preferred. I could have created feature branches for each improvement or feature, in the way I've done it, a branch has multiple and various features/improvements mixed. This could be improved in future projects.
- I haven't paid much attention to the repository configuration (intentionally due to lack of time), like adding branch protection rules, code owners, templates for issues and pull requests, etc. These are important for collaboration in teams and could be added later.
- API versioning wasn't considered in this implementation. This could be added using URL versioning (e.g., /v1/payments) or header-based versioning. The good thing is that the OpenAPI spec is already prepared for versioning, and is easy to implement.
- Notifications: Email/SMS notifications for payment status updates are not implemented. This could be implemented using topics in Kafka or RabbitMQ to send events to a notification service.
- Multi-Currency Support: While the system has a currency field, full multi-currency support, including exchange rates and conversions, is not included.
- Advanced Fraud Detection: Integration with fraud detection services to monitor and prevent fraudulent transactions is not part of the current implementation.
- User Authentication and Authorization: The service does not include user authentication or authorization mechanisms. This could be added using OAuth2, JWT, or integration with identity providers.
- Reporting and Analytics: There are no built-in reporting or analytics features for tracking payment trends, user behavior, or financial summaries.
- Mobile SDKs: Client libraries or SDKs for mobile platforms (iOS/Android) to facilitate integration with mobile applications are not provided. We could use a BFF (Backend For Frontend) pattern to create tailored APIs for mobile clients.
Or maybe create table views optimized for mobile usage, or web.
- Webhooks: Support for webhooks to notify external systems of payment events is not included.
- Admin Dashboard: An administrative dashboard for managing wallets, transactions, and monitoring system health is not implemented.
- Rate Limiting: There is no rate limiting on API endpoints to prevent abuse or excessive usage. We could implement this using API gateways or middleware, or some infrastructure like Redis as explained in the performance optimization doc, and Load Balancers if we deploy in cloud environments.
- Microservices Architecture: The current implementation is monolithic. A microservices architecture could be considered for better scalability and maintainability.
- CQRS Pattern: The application does not implement the Command Query Responsibility Segregation (CQRS) pattern, which could help in scaling read and write operations independently.
- GraphQL API: A GraphQL API alternative to the RESTful API is not provided.
- Other services, like users service, notification service, reporting service, etc.
- Security Enhancements: Additional security measures such as encryption at rest, enhanced logging, and monitoring are not fully implemented. We can add security for internal communication among other services using api keys.
- Secret Vaults: Integration with secret management solutions (e.g., AWS Secrets Manager, HashiCorp Vault) for managing sensitive configuration data is not included.
- CI/CD Pipeline: A complete continuous integration and continuous deployment (CI/CD) pipeline for automated, current workflow only checks builds, tests, linting and vulnerabilities.
- Another way of improve the CI/CD could be to sepparate each action from the workflow, so when a PR is created, only the necessary actions are run instead of all of them. And if only one fails, we can re-run only that one instead of the whole workflow.
- Consider some IAAS deployment options like Terraform or CloudFormation templates for easier infrastructure management.
- We could separate Dockerfiles for each service (API, workers, relay, mock gateway) for better modularity and independent scaling.
 
Notes:
I'm sure that if we sit and keep chatting  about this, other great features could come up, but these are the main ones that I think are worth mentioning for now (⌐ ͡■ ͜ʖ ͡■).

My intention was to add a section with lectures and learned lessons during the development of this project, but due to time constraints I couldn't include it. This could be a valuable addition for future reference. I might update this in the future, and maybe pour all this information in a Notion page for easier access and organization.