# Spec 113: Deep-Linking & Share URLs

## Status
done

## Overview
Users can right-click a view or business domain and select "Share (copy URL)..." to copy a shareable deep-link. Opening the link navigates directly to that view or domain.

## Requirements

### Context Menu
- Add "Share (copy URL)..." item to view context menu in Navigation Tree
- Add "Share (copy URL)..." item to domain context menu in Business Domains sidebar
- Available to all users (read-only action, no permission check)

### URL Format
- Architecture Views: `/?view={viewId}`
- Business Domains: `/business-domains?domain={domainId}`
- URLs are absolute (include origin)

### Copy Behavior
- Use Clipboard API to copy generated URL
- Show success toast "Link copied to clipboard" using existing react-hot-toast
- Show error toast on clipboard failure

### Deep-Link Navigation
- On app initialization, parse `view` query parameter and navigate to that view
- On Business Domains page load, parse `domain` query parameter and visualize that domain
- Clean query parameters from URL after processing

### Error Handling
- If linked view doesn't exist: show error toast, redirect to default view
- If linked domain doesn't exist: show error toast, show domain list

## Checklist
- [x] Add share menu item to view context menu (useTreeContextMenus)
- [x] Add share menu item to domain context menu (useDomainContextMenu)
- [x] Create clipboard utility function
- [x] Parse view parameter in useAppInitialization
- [x] Parse domain parameter in BusinessDomainsPage
- [x] Handle invalid deep-links with error toast and redirect
- [x] Clean URL after navigation
- [x] Tests passing
- [x] User sign-off
