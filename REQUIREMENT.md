Software Architect/Engineering Manager Take Home Assignment
We are excited to invite you to the technical interview stage. As a Software Architect/Engineering Manager at Durianpay, you will play a key role in designing, building, and maintaining scalable systems that power our platform.

This take-home assignment is your opportunity to demonstrate your approach to backend design, code quality, and communication.

1. Background
   Your task is to design and implement a Configuration Management Service.
   This service will allow teams to safely define, update, and retrieve configuration data, with schema-based validation, versioning, and rollback support.

We are evaluating your approach to API design, validation, code structure, and documentation.

2. Requirements
   2.1 Functional Requirements
   Please implement the following API endpoints (RESTful preferred):

2.1.1 Create a Configuration
Accepts a config name and its data (JSON).
Validates the config against a schema (specific to config type).
Stores it with version 1.
2.1.2 Update a Configuration
Accepts an updated JSON payload.
Validates the update using the same schema.
Increments the version number.
2.1.3 Rollback a Configuration
Allows reverting to any previous version.
Creates a new version by applying the chosen historical data.
2.1.4 Fetch Config
Retrieves by config name (returns latest version).
Optional: Retrieves a specific version.
2.1.5 List Versions
Lists all historical versions for a given config.
2.2 Non-Functional Requirements
2.2.1 Technical Requirements
Programming Language: Golang only
Framework: Any standard Go web framework
Persistence: Choose from in-memory (Map), file-based (JSON), or SQLite
Config Schema: Hardcode schemas in the service (using JSON Schema or structs)
API Style: RESTful preferred
Validation: Ensure all config JSONs are validated against their schema before storing or updating
Design Flexibility: Design your system architecture to accommodate various types of configurations
Robustness: Handle unexpected scenarios gracefully and provide useful troubleshooting information
Production Readiness: Design your implementation to be effective in real-world, multi-user environments
Example Schema for payment_config:

json


1
2
3
4
5
6
7
8
⌄
⌄
{
"type": "object",
"properties": {
"max_limit": { "type": "integer" },
"enabled": { "type": "boolean" }
},
"required": ["max_limit", "enabled"]
}
Example Input:

json


1
2
3
4
⌄
{
"max_limit": 1000,
"enabled": true
}
2.2.2 Project Setup & API Definition
Implement your solution in Golang only.
Structure your project with at least a main.go, README.md, and a Makefile.
Define your API contract in an OpenAPI v3 specification (e.g., api.yml or openapi.yaml) at the project root.
The API spec should describe all endpoints, request/response bodies, status codes, error structures, and provide example payloads.
Ensure your implementation and API definition are aligned.
2.2.3 Testing and Code Quality
We highly value:

Well-organized, readable code with clear documentation and descriptive variable names.
Comprehensive testing, including both unit and functional tests.
Code that follows Go formatting and style guidelines.
3. Submission Guidelines & Review Mechanism
   3.1 Deliverables
   Provide a working service with clear setup instructions. Include:

Comprehensive test cases covering CRUD operations, validation, and rollback functionality.
A README.md that includes:
Complete instructions for setting up and running your application.
API documentation (preferably in OpenAPI/Swagger format).
Schema explanation.
Notes on your design decisions and trade-offs.
Feel free to include ideas for improvements, additional features, or creative solutions beyond the listed requirements.
🚨 Reminder: This assignment should be completed within 48 hours of receiving it by email.

3.2 How to Submit
Upload your code to a new GitHub repository (public or private).
If your repository is private, add the GitHub account durianpay-tech-interviewer as a collaborator so we can review your work.
(If you have difficulties adding the above GitHub account, you can add gautamdurian or jhoviedp instead.)
No starter template or project skeleton is provided; structure the project as you think is best.
3.3 Reviewer Environment
Your submission will be reviewed on a machine with:

Go v1.21+
Docker & Docker Compose
Makefile
Mac-based computer
Please ensure your code and instructions are compatible with this environment.

3.4 Validating The Test
Your submission must include a README.md that clearly documents:

All prerequisites and dependencies.
Step-by-step instructions for setting up the environment.
Exact commands to build, start, and test your project (unit and API/integration).
💡 Note: Submissions that fail to execute properly following your instructions in our reviewer environment will be automatically rejected. We recommend testing your setup instructions on a clean environment before submission.

3.5 Post-Submission Process
After you submit your assignment, our technical team will review your code within 2 to 5 business days.
You will receive an email with feedback and next steps regardless of the outcome.

For any technical questions during the assignment, please contact tech-recruitment@durianpay.id.

4. Evaluation Criteria
   Your code will be evaluated based on several key factors:

Functional Completeness: Implementation of all required features.
Code Quality: Organization, error handling, and Go best practices.
API Design: Intuitive, documented RESTful endpoints.
Testing: Comprehensive test coverage.
Documentation: Clear explanations of design decisions.
We will also evaluate how you handle edge cases, performance considerations, and your approach to validation and versioning. We value clear communication in your documentation and comments.

Remember, this assignment showcases not just your coding abilities but also how you think about software design and communicate your decisions. We encourage you to demonstrate your creativity and problem-solving skills while meeting the requirements.

We look forward to seeing your submission. Good luck!