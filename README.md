# Chirpy (Twitter Clone)
- Built with the following:
    - Only Golang STD library and Raw SQL
    - No backend framework
    - No ORM
    - Implemented JWT with SigningMethodHS256 ( has both features regarding access tokens and refresh tokens )
    - Implemented Argon2id for hashing user passwords before entering the database
    - Simple sorting based on when chirps (Tweets) were created at (ascending / descending)
    - Made use of private/public keys for validation via JWT and Webhooks
## RESTFUL APIs Architecture
- Basic Crud functionality
- Monolithic Approach
### Tools
- PSQL (postgres terminal)
- Goose (migration tool for databases)
- SQLC (Compile Raw SQL into type-safe code)
