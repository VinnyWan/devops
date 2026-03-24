# Fix Authentication and Menu Issues Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix three critical issues: 404 on permissions endpoint, auto-logout on page refresh, and sidebar menu auto-collapse behavior.

**Architecture:** Backend API method mismatch fix, frontend session persistence improvement, and NMenu accordion configuration.

**Tech Stack:** Go (Gin), TypeScript (Vue 3), Naive UI

---

## Issue 1: Fix 404 on /api/v1/user/permissions Endpoint

**Root Cause:** Frontend sends POST request but backend registers GET endpoint.

### Task 1.1: Update Backend Route to POST

**Files:**
- Modify: `backend/routers/v1/user.go:14`

**Step 1: Change GET to POST for permissions endpoint**

```go
r.POST("/user/permissions", api.GetUserPermissions)
```

**Step 2: Verify route registration**

Run: `cd backend && go build ./cmd/server`
Expected: Build succeeds without errors

**Step 3: Commit**

```bash
git add backend/routers/v1/user.go
git commit -m "fix: change user permissions endpoint from GET to POST"
```

---

## Issue 2: Fix Auto-Logout on Page Refresh

**Root Cause:** Session validation fails when fetching permissions, causing clearAuth() to be called.

### Task 2.1: Add Error Handling to Prevent Auto-Logout

**Files:**
- Modify: `frontend/src/router/index.ts:31-37`

**Step 1: Update permission fetch error handling**

Replace lines 31-37 with:

```typescript
  if (!authStore.permissionsLoaded) {
    try {
      await authStore.fetchPermissions()
    } catch (error) {
      console.error('Failed to fetch permissions:', error)
      // Don't logout on permission fetch failure, just skip permission check
      authStore.permissionsLoaded = true
    }
  }
```

**Step 2: Test the change**

1. Start dev server: `cd frontend && npm run dev`
2. Login to the application
3. Refresh the page (F5)
4. Expected: User remains logged in even if permissions fail to load

**Step 3: Commit**

```bash
git add frontend/src/router/index.ts
git commit -m "fix: prevent auto-logout on permission fetch failure during page refresh"
```

### Task 2.2: Improve Session Cookie Configuration

**Files:**
- Modify: `backend/internal/modules/user/api/auth.go:27-31`

**Step 1: Update session cookie settings for better persistence**

Replace lines 27-31 with:

```go
func setSessionCookie(c *gin.Context, value string, maxAge int) {
	secure := config.Cfg.GetString("server.mode") == "release"
	c.SetSameSite(http.SameSiteLaxMode)
	// Set path to "/" to ensure cookie is sent with all API requests
	c.SetCookie("session_id", value, maxAge, "/", "", secure, true)
}
```

**Step 2: Verify cookie is set correctly**

1. Login via API
2. Check browser DevTools > Application > Cookies
3. Expected: session_id cookie with Path="/" and HttpOnly=true

**Step 3: Commit**

```bash
git add backend/internal/modules/user/api/auth.go
git commit -m "fix: ensure session cookie path is set to root"
```

---

## Issue 3: Enable Sidebar Menu Auto-Collapse

**Root Cause:** NMenu component missing `accordion` prop to enable auto-collapse behavior.

### Task 3.1: Add Accordion Mode to Menu

**Files:**
- Modify: `frontend/src/layouts/MainLayout.vue:130-137`

**Step 1: Add accordion prop to NMenu**

Replace lines 130-137 with:

```vue
          <n-menu
            :collapsed="collapsed"
            :collapsed-width="64"
            :collapsed-icon-size="22"
            :options="menuOptions"
            :value="activeKey"
            :accordion="true"
            @update:value="handleMenuUpdate"
          />
```

**Step 2: Test menu behavior**

1. Start dev server: `cd frontend && npm run dev`
2. Login and navigate to main layout
3. Click "容器管理" to expand
4. Click "系统管理" to expand
5. Expected: "容器管理" automatically collapses when "系统管理" expands

**Step 3: Commit**

```bash
git add frontend/src/layouts/MainLayout.vue
git commit -m "feat: enable accordion mode for sidebar menu auto-collapse"
```

---

## Verification

### Task 4.1: End-to-End Testing

**Step 1: Start backend server**

```bash
cd backend
go run cmd/server/main.go
```

**Step 2: Start frontend dev server**

```bash
cd frontend
npm run dev
```

**Step 3: Test all three fixes**

1. **Test permissions endpoint:**
   - Login with valid credentials
   - Open DevTools > Network tab
   - Check for POST request to `/api/v1/user/permissions`
   - Expected: Status 200 (not 404)

2. **Test page refresh:**
   - After login, refresh the page (F5)
   - Expected: User remains logged in, no redirect to login page

3. **Test menu auto-collapse:**
   - Click "容器管理" menu item
   - Click "系统管理" menu item
   - Expected: "容器管理" collapses automatically

**Step 4: Final commit**

```bash
git add -A
git commit -m "docs: add verification notes for auth and menu fixes"
```

---

## Rollback Plan

If issues occur:

```bash
git revert HEAD~3..HEAD
```

This reverts all three commits in reverse order.
