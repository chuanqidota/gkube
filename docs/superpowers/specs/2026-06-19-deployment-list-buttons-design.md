# Deployment List Button Simplification Design

**Date:** 2026-06-19
**Status:** Draft
**Author:** Claude (Brainstorming)

## Problem Statement

The current Deployment list page has 4 row action buttons (YAML, Scale, Restart, Delete) taking 320px width — the widest among all list pages. This is inconsistent with other list pages (200px) and may feel cluttered.

## Decision

Simplify the Deployment list row actions to match other list pages, moving Deployment-specific operations (Scale, Restart) to the detail page only.

## Design

### List Page Changes

**Row Actions (simplified):**
- YAML — View/Edit deployment YAML
- Delete — Delete deployment with confirmation

**Action Column Width:** 200px (consistent with StatefulSet, DaemonSet, etc.)

**Toolbar (unchanged):**
- Search by name
- Namespace filter
- Refresh (with countdown)
- Auto-refresh toggle
- Create button
- Batch Delete button

### Detail Page Additions

Scale and Restart operations will be available in the Deployment detail page:

- **Scale button** — Opens dialog to adjust replicas
- **Restart button** — Triggers rolling restart with confirmation

These should be placed in the detail page header/action area alongside other actions.

### Visual Layout

**Before (current):**
```
│ 操作 (320px)                    │
│ [YAML] [伸缩] [重启] [删除]     │
```

**After (simplified):**
```
│ 操作 (200px)    │
│ [YAML] [删除]   │
```

## Rationale

1. **Consistency** — All list pages should have similar action column widths
2. **Cleaner UI** — List page focuses on viewing and filtering, not operations
3. **Detail page is appropriate** — Scale and Restart are less frequent operations that benefit from the context provided in the detail view
4. **Industry practice** — Rancher follows this pattern (Scale in detail page)

## Scope

- Modify `DeploymentList.vue` to remove Scale and Restart row actions
- Adjust action column width from 320px to 200px
- Ensure Scale and Restart buttons exist in Deployment detail page (if not already present)

## Out of Scope

- Changes to other list pages
- Adding "more actions" dropdown or context menu
- Changes to toolbar buttons
