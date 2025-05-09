Possible Solutions
1. Custom Policy Approach

Implement a custom policy using the Identity Experience Framework (IEF)
Create a verification step that checks against a pre-approved list
Require additional verification information like a registration code or child's ID

2. Pre-Generated Accounts

Create accounts in advance for expected parents
Send temporary/one-time passwords to verified parents
Require password change on first login
Benefit: Complete control over who can access the system

3. API-Connected User Flow

Extend the standard user flow with an API connector
Validate registration against your database of approved families
Block account creation for unauthorized users during registration
Benefit: Automated but controlled access

4. Invitation-Only System

Implement Azure AD B2C custom policies with invitation codes
Send unique invitation links to parents
Use Microsoft Graph API to automate the invitation process
