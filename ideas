package main

import (
    "crypto/rand"
    "crypto/rsa"
    "crypto/tls"
    "crypto/x509"
    "encoding/base64"
    "encoding/json"
    "encoding/pem"
    "fmt"
    "io/ioutil"
    "log"
    "net/smtp"
    "net/url"
    "os"
    "strings"
    "time"
)

// Configuration for the invitation system
type Config struct {
    // Azure B2C Configuration
    TenantName      string
    TenantDomain    string // usually yourtenantname.onmicrosoft.com
    ClientID        string
    RedirectURI     string
    PolicyName      string
    CertificatePath string
    
    // SMTP Configuration
    SMTPHost     string
    SMTPPort     int
    SMTPUsername string
    SMTPPassword string
    SenderEmail  string
    SenderName   string
}

// JWT token structure for Azure B2C invitation
type JWTHeader struct {
    Alg string `json:"alg"`
    Kid string `json:"kid"`
    Typ string `json:"typ"`
}

type JWTPayload struct {
    Iss string `json:"iss"`
    Sub string `json:"sub"`
    Aud string `json:"aud"`
    Exp int64  `json:"exp"`
    Nbf int64  `json:"nbf"`
    Iat int64  `json:"iat"`
    
    // Custom claims
    Email string `json:"email"`
}

// Load RSA private key from PEM file
func loadPrivateKey(filename string) (*rsa.PrivateKey, error) {
    keyPEM, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("failed to read certificate file: %v", err)
    }
    
    block, _ := pem.Decode(keyPEM)
    if block == nil {
        return nil, fmt.Errorf("failed to parse PEM block containing the key")
    }
    
    privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
    if err != nil {
        return nil, fmt.Errorf("failed to parse private key: %v", err)
    }
    
    return privateKey, nil
}

// Create a JWT token for invitation
func createInvitationToken(email string, config Config) (string, error) {
    // Load private key
    privateKey, err := loadPrivateKey(config.CertificatePath)
    if err != nil {
        return "", err
    }
    
    // Create JWT header
    header := JWTHeader{
        Alg: "RS256",
        Kid: "invitation-signing-key", // This should match your key ID in Azure B2C
        Typ: "JWT",
    }
    
    // Current time for token timing
    now := time.Now()
    
    // Create JWT payload
    payload := JWTPayload{
        Iss: config.ClientID,
        Sub: email,
        Aud: config.ClientID,
        Nbf: now.Unix(),
        Iat: now.Unix(),
        Exp: now.Add(24 * time.Hour).Unix(), // 24-hour expiry
        
        // Custom claims
        Email: email,
    }
    
    // Convert header to JSON and encode
    headerJSON, err := json.Marshal(header)
    if err != nil {
        return "", fmt.Errorf("error encoding JWT header: %v", err)
    }
    headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)
    
    // Convert payload to JSON and encode
    payloadJSON, err := json.Marshal(payload)
    if err != nil {
        return "", fmt.Errorf("error encoding JWT payload: %v", err)
    }
    payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadJSON)
    
    // Create signature input (header + "." + payload)
    signatureInput := headerEncoded + "." + payloadEncoded
    
    // Sign the token
    signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, x509.SHA256WithRSA, []byte(signatureInput))
    if err != nil {
        return "", fmt.Errorf("error signing JWT: %v", err)
    }
    signatureEncoded := base64.RawURLEncoding.EncodeToString(signature)
    
    // Create complete JWT token
    token := signatureInput + "." + signatureEncoded
    
    return token, nil
}

// Generate a B2C invitation URL with the JWT token
func generateInvitationURL(email string, config Config) (string, error) {
    // Create JWT token
    token, err := createInvitationToken(email, config)
    if err != nil {
        return "", fmt.Errorf("error creating invitation token: %v", err)
    }
    
    // Base URL for Azure B2C
    baseURL := fmt.Sprintf(
        "https://%s.b2clogin.com/%s/%s/oauth2/v2.0/authorize",
        config.TenantName,
        config.TenantDomain,
        config.PolicyName,
    )
    
    // Build query parameters
    params := url.Values{}
    params.Add("client_id", config.ClientID)
    params.Add("nonce", fmt.Sprintf("defaultNonce-%d", time.Now().Unix()))
    params.Add("redirect_uri", config.RedirectURI)
    params.Add("scope", "openid profile")
    params.Add("response_type", "id_token")
    params.Add("response_mode", "form_post")
    params.Add("prompt", "login")
    params.Add("id_token_hint", token)
    
    // Combine URL and parameters
    fullURL := baseURL + "?" + params.Encode()
    
    return fullURL, nil
}

// Send invitation email using net/smtp
func sendInvitationEmail(email, invitationURL string, config Config) error {
    // SMTP server address
    smtpAddress := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)
    
    // Email content
    from := fmt.Sprintf("%s <%s>", config.SenderName, config.SenderEmail)
    to := email
    subject := "Invitation to join our application"
    
    // Generate a boundary for MIME parts
    boundary := fmt.Sprintf("boundary-%d", time.Now().UnixNano())
    
    // Create email headers
    headers := []string{
        fmt.Sprintf("From: %s", from),
        fmt.Sprintf("To: %s", to),
        fmt.Sprintf("Subject: %s", subject),
        fmt.Sprintf("MIME-Version: 1.0"),
        fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s", boundary),
        "",
    }
    
    // Create plain text version
    plainText := fmt.Sprintf(
        "You've been invited to join our application.\n\n" +
        "Please visit this link to sign up: %s\n\n" +
        "This invitation link will expire in 24 hours.",
        invitationURL,
    )
    
    // Create HTML version with better formatting
    htmlBody := fmt.Sprintf(`
    <html>
    <head>
        <style>
            body { font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; }
            .header { background-color: #0078d4; padding: 20px; text-align: center; color: white; }
            .content { padding: 20px; }
            .button { 
                display: inline-block; 
                background-color: #0078d4; 
                color: white; 
                padding: 12px 24px; 
                text-decoration: none; 
                border-radius: 4px; 
                font-weight: bold; 
                margin: 20px 0;
            }
            .footer { padding: 20px; font-size: 12px; color: #666; text-align: center; }
            .link-text { word-break: break-all; color: #0078d4; }
        </style>
    </head>
    <body>
        <div class="header">
            <h2>You're Invited!</h2>
        </div>
        <div class="content">
            <p>You've been invited to create an account in our application.</p>
            <p>Click the button below to set up your account.</p>
            
            <div style="text-align: center;">
                <a href="%s" class="button">Create Your Account</a>
            </div>
            
            <p>If the button doesn't work, copy and paste this link in your browser:</p>
            <p class="link-text">%s</p>
            
            <p><strong>Note:</strong> This invitation link will expire in 24 hours.</p>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply to this email.</p>
        </div>
    </body>
    </html>`, invitationURL, invitationURL)
    
    // Build the complete message with both text and HTML parts
    var messageBody strings.Builder
    
    // Add headers
    messageBody.WriteString(strings.Join(headers, "\r\n"))
    
    // Add plain text part
    messageBody.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
    messageBody.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
    messageBody.WriteString(plainText)
    
    // Add HTML part
    messageBody.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
    messageBody.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
    messageBody.WriteString(htmlBody)
    
    // Close boundary
    messageBody.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))
    
    // Connect to the SMTP server
    client, err := smtp.Dial(smtpAddress)
    if err != nil {
        return fmt.Errorf("failed to connect to SMTP server: %v", err)
    }
    defer client.Close()
    
    // Start TLS
    if err = client.StartTLS(&tls.Config{ServerName: config.SMTPHost}); err != nil {
        return fmt.Errorf("failed to start TLS: %v", err)
    }
    
    // Authenticate with the server
    auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPHost)
    if err = client.Auth(auth); err != nil {
        return fmt.Errorf("SMTP authentication failed: %v", err)
    }
    
    // Set sender and recipient
    if err = client.Mail(config.SenderEmail); err != nil {
        return fmt.Errorf("failed to set sender: %v", err)
    }
    
    if err = client.Rcpt(to); err != nil {
        return fmt.Errorf("failed to set recipient: %v", err)
    }
    
    // Send the email body
    writer, err := client.Data()
    if err != nil {
        return fmt.Errorf("failed to open data stream: %v", err)
    }
    
    _, err = writer.Write([]byte(messageBody.String()))
    if err != nil {
        return fmt.Errorf("failed to write email data: %v", err)
    }
    
    if err = writer.Close(); err != nil {
        return fmt.Errorf("failed to close data stream: %v", err)
    }
    
    // Close the connection
    client.Quit()
    
    return nil
}

// Helper function to create a self-signed certificate for development
func createSelfSignedCert(filename string) error {
    // Generate private key
    privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
    if err != nil {
        return fmt.Errorf("failed to generate private key: %v", err)
    }
    
    // Encode private key to PEM
    keyPEM := &pem.Block{
        Type:  "RSA PRIVATE KEY",
        Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
    }
    
    // Write to file
    keyFile, err := os.Create(filename)
    if err != nil {
        return fmt.Errorf("failed to create key file: %v", err)
    }
    defer keyFile.Close()
    
    if err := pem.Encode(keyFile, keyPEM); err != nil {
        return fmt.Errorf("failed to write key to file: %v", err)
    }
    
    return nil
}

// Main invitation function
func sendInvitation(email string, config Config) error {
    // Generate invitation URL with JWT token
    invitationURL, err := generateInvitationURL(email, config)
    if err != nil {
        return fmt.Errorf("failed to generate invitation URL: %v", err)
    }
    
    // Send invitation email
    err = sendInvitationEmail(email, invitationURL, config)
    if err != nil {
        return fmt.Errorf("failed to send invitation email: %v", err)
    }
    
    return nil
}

func main() {
    // Load configuration
    config := Config{
        // Azure B2C Configuration
        TenantName:      "yourtenant",
        TenantDomain:    "yourtenant.onmicrosoft.com",
        ClientID:        "your-client-id",
        RedirectURI:     "https://your-app.com/auth-callback",
        PolicyName:      "B2C_1A_SignUpInvitation", // Your custom policy for invitation
        CertificatePath: "invitation-key.pem",      // Path to your private key file
        
        // SMTP Configuration
        SMTPHost:     "smtp.gmail.com",
        SMTPPort:     587,
        SMTPUsername: "your-email@gmail.com",
        SMTPPassword: "your-app-password",     // App password for Gmail
        SenderEmail:  "your-email@gmail.com",
        SenderName:   "Your Application",
    }
    
    // Check if certificate exists, if not, create one for development
    if _, err := os.Stat(config.CertificatePath); os.IsNotExist(err) {
        log.Println("Certificate not found, creating self-signed certificate for development...")
        err := createSelfSignedCert(config.CertificatePath)
        if err != nil {
            log.Fatalf("Failed to create certificate: %v", err)
        }
        log.Println("Certificate created successfully. For production, replace with a proper certificate.")
    }
    
    // Example: Send invitation to a user
    userEmail := "newuser@example.com"
    log.Printf("Sending invitation to %s...", userEmail)
    
    err := sendInvitation(userEmail, config)
    if err != nil {
        log.Fatalf("Failed to send invitation: %v", err)
    }
    
    log.Printf("Invitation sent successfully to %s", userEmail)
}

// For HTTP server integration, you might want to add an HTTP handler:
/*
func inviteHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Parse form data
    err := r.ParseForm()
    if err != nil {
        http.Error(w, "Failed to parse form", http.StatusBadRequest)
        return
    }
    
    // Get email from form
    email := r.FormValue("email")
    if email == "" {
        http.Error(w, "Email is required", http.StatusBadRequest)
        return
    }
    
    // Load configuration
    config := Config{
        // Your configuration here...
    }
    
    // Send invitation
    err = sendInvitation(email, config)
    if err != nil {
        log.Printf("Error sending invitation: %v", err)
        http.Error(w, "Failed to send invitation", http.StatusInternalServerError)
        return
    }
    
    // Respond with success
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"success","message":"Invitation sent successfully"}`))
}
*/
