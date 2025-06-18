# Auth0 Setup Guide for Admin Permissions

This guide explains how to configure Auth0 so your application can restrict access to the `/admin` page using Auth0 permissions.

## 1. Create an API in Auth0
1. Go to your Auth0 dashboard.
2. Navigate to **Applications > APIs**.
3. Click **Create API**.
4. Set:
   - **Name**: (e.g.) `SchoolScienceHelper API`
   - **Identifier**: (e.g.) `https://yourdomain/api` (must match your backend's audience)
   - **Signing Algorithm**: RS256
5. Click **Create**.

## 2. Define Permissions
1. In your API settings, go to the **Permissions** tab.
2. Click **Add Permission**.
3. Add a permission for admin access, e.g.:
   - **Name**: `admin:panel`
   - **Description**: `Access the admin panel`
4. Save.

## 3. Create an Admin Role
1. Go to **User Management > Roles**.
2. Click **Create Role**.
3. Name it (e.g.) `Admin`.
4. Save.

## 4. Assign Permissions to the Role
1. Click your new `Admin` role.
2. Go to the **Permissions** tab.
3. Click **Assign Permissions**.
4. Select your API and the `admin:panel` permission.
5. Save.

## 5. Assign the Role to Users
1. Go to **User Management > Users**.
2. Click a user you want to make admin.
3. Go to the **Roles** tab.
4. Click **Assign Roles** and select `Admin`.

## 6. Ensure Permissions are Included in Tokens
1. Go to **Applications > APIs** and select your API.
2. In the **Machine to Machine Applications** tab, authorize your application if needed.
3. By default, Auth0 includes permissions in the `permissions` claim of the access token.

If you use rules or actions to customize tokens, ensure the `permissions` array is present in the token.

## 7. Configure Your Backend
- Set the environment variable in your backend:
  ```sh
  export ADMIN_PERMISSION=admin:panel
  ```
- Your backend will now only allow users with this permission to access `/admin`.

---

**Summary:**
- Create an API and permission in Auth0.
- Assign the permission to a role.
- Assign the role to users.
- Set `ADMIN_PERMISSION` in your backend to match the permission string.
- Only users with that permission can access the admin page.
