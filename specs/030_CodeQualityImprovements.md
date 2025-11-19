# Code quality improvements
This spec lists files that are subject to code quality or security improvements

## Refactoring targets
### /backend/internal/architecturemodeling/infrastructure/api/relation_handlers_integration_test.go
- createTestRelation has excess number of arguments (7) should be max 4.
- TestGetRelationsFromComponent_Integration and TestGetRelationsToComponent_Integration has code duplication
- TestCascadeDeleteRelations_Integration is too large (111 lines)
### /frontend/src/store&appStore.ts
- The module contains 4 functions with similar structure:
-- updateComponent
-- updateRelation
-- setEdgeType
-- setLayoutDirection

A certain degree of duplicated code might be acceptable. The problems start when it is the same behavior that is duplicated across the functions in the module, ie. a violation of the Don't Repeat Yourself (DRY) principle. DRY violations lead to code that is changed together in predictable patterns, which is both expensive and risky. 
Once you have identified the similarities across functions, look to extract and encapsulate the concept that varies into its own function(s). These shared abstractions can then be re-used, which minimizes the amount of duplication and simplifies change.

- All functions in the file has primitive obsession

- All functions in the file has String Heavy Function Arguments
String is a generic type that fail to capture the constraints of the domain object it represents. In this module, 71 % of all function arguments are string types.
Heavy string usage indicates a missing domain language. Introduce data types that encapsulate the semantics. For example, a user_name is better represented as a constrained User type rather than a pure string, which could be anything.

### /backend/internal/architectureviews/application/readmodels/architecture_view_read_model.go
- Bumpy Road: This file has 1 bumpy roads in GetAll(bumps = 2)
A Bumpy Road is a function that contains multiple chunks of nested conditional logic inside the same function. The deeper the nesting and the more bumps, the lower the code health.

A bumpy code road represents a lack of encapsulation which becomes an obstacle to comprehension. In imperative languages there’s also an increased risk for feature entanglement, which leads to complex state management.
Bumpy Road implementations indicate a lack of encapsulation.
A Bumpy Road often suggests that the function/method does too many things. The first refactoring step is to identify the different possible responsibilities of the function. Consider extracting those responsibilities into smaller, cohesive, and well-named functions. The EXTRACT FUNCTION refactoring is the primary response.

- Excess Number of Function Arguments for two functions
- Primitive Obsession generally in file

### /backend/internal/architectureviews/infrastructure/api/view_handlers.go

### /backend/internal/architectureviews/domain/aggregates/architecture_view.go
- apply is a complex method that is called often.
- code duplication occurs in many functions

## Security issues
SQL Injection	backend/internal/infrastructure/migrations/runner.go	57:13 - 57:31
SQL Injection	backend/internal/infrastructure/migrations/runner.go	145:15 - 145:22
Path Traversal	backend/internal/infrastructure/migrations/runner.go	57:13 - 57:31
Path Traversal	backend/internal/infrastructure/migrations/runner.go	132:18 - 132:29
Container or Pod is running without root user control. You can resolve it by Set securityContext.runAsNonRoot to true	k8s/frontend-deployment.yaml		
Container or Pod is running without root user control	k8s/backend-deployment.yaml		
Container or Pod is running without root user control	k8s/backend-deployment.yaml		
Container is running without privilege escalation control	k8s/backend-deployment.yaml		
Container is running without privilege escalation control	k8s/frontend-deployment.yaml		
Container is running without privilege escalation control	k8s/backend-deployment.yaml
Container does not drop all default capabilities	k8s/backend-deployment.yaml		
Container does not drop all default capabilities	k8s/frontend-deployment.yaml		
Container does not drop all default capabilities	k8s/backend-deployment.yaml		
Container's or Pod's UID could clash with host's UID	k8s/backend-deployment.yaml		
Container's or Pod's UID could clash with host's UID	k8s/backend-deployment.yaml		
Container's or Pod's UID could clash with host's UID	k8s/frontend-deployment.yaml		
Container is running with writable root filesystem	k8s/backend-deployment.yaml		
Container is running with writable root filesystem	k8s/frontend-deployment.yaml		
Container is running with writable root filesystem	k8s/backend-deployment.yaml		
Container could be running with outdated image	k8s/frontend-deployment.yaml
Container could be running with outdated image	k8s/backend-deployment.yaml		
Container could be running with outdated image	k8s/backend-deployment.yaml		
Open Redirect	backend/go.mod Introduced through github.com/go-chi/chi/v5/middleware@5.0.11 Fixed in github.com/go-chi/chi/v5/middleware@5.2.2

To fix sql injection:
Use Parameterized Queries: Always use prepared statements or parameter placeholders ($1, $2, etc.) instead of building SQL strings directly – use it consistently.
Input Sanitization and Validation: Rigorously validate and sanitize inputs in your application. Reject or cleanse anything that doesn’t meet criteria (e.g., length limits, allowed characters). This reduces the chance that dangerous payloads even reach your query layer.
No Direct String Concatenation: Rarely if ever concatenate raw input into an SQL command. If you absolutely must construct SQL dynamically (for example, dynamic ORDER BY or table names), use safe helper functions. For instance, in Node you might use the pg-format library (which properly escapes identifiers) for those rare cases, or in SQL use format() with %I for identifiers and %L for literals.

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests updated if relevant
- [ ] User sign-off
