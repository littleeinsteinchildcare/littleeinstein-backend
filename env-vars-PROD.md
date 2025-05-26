```markdown
# 🔐 Step-by-Step Guide to Configure Azure Key Vault References in App Service

This guide walks you through securely injecting secrets into your application running on **Azure App Service** using **Azure Key Vault references**.

---

## 📘 1. Access the Azure Portal

- Log in to the [Azure Portal](https://portal.azure.com)
- Navigate to your **App Service**
- Click on the **App Service name** to open its management blade

---

## ⚙️ 2. Open Configuration Settings

- In the left-hand menu of your App Service, click on **Configuration**
- This opens the **Application settings** page
- These settings are used to define environment variables available to your app

---

## ➕ 3. Add an Application Setting

- Click on **+ New application setting**
- Set the **Name** to match the environment variable your code expects  
  Example: `DB_PASSWORD`
- Your Go code can access it using:

  ```go
  os.Getenv("DB_PASSWORD")
  ```

---

## 🔑 4. Set the Key Vault Reference

- In the **Value** field, use the following syntax:

  ```
  @Microsoft.KeyVault(SecretUri=https://<your-vault-name>.vault.azure.net/secrets/<your-secret-name>/)
  ```

- Replace `<your-vault-name>` and `<your-secret-name>` with your actual Key Vault and secret name

---

## 💾 5. Save the Configuration

- Click **OK** to add the setting
- Click **Save** at the top of the Configuration page
- Azure will **restart your App Service** to apply the changes

---

## 🧠 How It Works Behind the Scenes

When your App Service starts:

1. Azure detects the **Key Vault reference syntax**
2. It uses the **App Service's managed identity** to authenticate with Key Vault
3. It retrieves the **secret value**
4. It injects the **resolved value** as an environment variable
5. Your Go app can read it using `os.Getenv()`

---

## ✅ Benefits

- 🔒 **No secret values hardcoded or stored in code**
- 🔐 **Secure authentication via Managed Identity**
- 🧼 **Cleaner code — no need to directly interact with Key Vault from your app**
```
