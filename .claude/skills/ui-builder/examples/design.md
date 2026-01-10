# Contact Manager with Chat

## Intent
Manage contacts with list/detail view. Search and filter. Chat with agent for assistance.

## Layout
```
┌─────────────────────────────────────────────────────┐
│ [🔍 Search contacts...        ] [3] [+ Add] [Dark]  │
├───────────────────────┬─────────────────────────────┤
│   Alice Smith         │ Name: [Alice Smith      ]   │
│ ▎Bob Jones      ←     │ Email: [bob@example.com ]   │
│   Carol White         │ Status: [Active ▼]          │
│                       │ VIP: [✓]                    │
│                       │ ─────────────────────────── │
│                       │ [Delete] [Cancel]    [Save] │
├───────────────────────┴─────────────────────────────┤
│ Chat with Agent                                     │
│ ┌─────────────────────────────────────────────────┐ │
│ │ Agent: How can I help you?                      │ │
│ │ You: Add a contact for John                     │ │
│ │ Agent: Done! I added John to your contacts.     │ │
│ └─────────────────────────────────────────────────┘ │
│ [Type a message...                    ] [Send]      │
└─────────────────────────────────────────────────────┘
```
## Components

| Element       | Binding                                   | Notes                     |
|---------------|-------------------------------------------|---------------------------|
| Search input  | ui-value="searchQuery?keypress"           | Live filter               |
| Count badge   | ui-value="contactCount()"                 | Shows filtered count      |
| Add btn       | ui-action="add()"                         | Creates new contact       |
| Dark toggle   | ui-value="darkMode"                       | sl-switch                 |
| Contact list  | ui-view="contacts()?wrapper=lua.ViewList" | Computed filtered list    |
| Row click     | ui-action="selectMe()"                    | Selects contact           |
| Row highlight | ui-class-selected="isSelected()"          | Shows selection state     |
| Detail panel  | ui-class-hidden="hideDetail"              | Hidden when no selection  |
| Name input    | ui-value="current.name"                   |                           |
| Email input   | ui-value="current.email"                  |                           |
| Status select | ui-value="current.status"                 | active/inactive           |
| VIP switch    | ui-value="current.vip"                    |                           |
| Delete btn    | ui-action="deleteCurrent()"               | variant="danger"          |
| Cancel btn    | ui-action="cancel()"                      | Discards changes          |
| Save btn      | ui-action="save()"                        | Inserts or updates        |
| Chat messages | ui-view="messages?wrapper=lua.ViewList"   |                           |
| Chat input    | ui-value="chatInput?keypress"             | Live input                |
| Send btn      | ui-action="sendChat()"                    | Fires pushState           |

## Behavior
- Type in search → filters contacts list in real-time
- Add → creates temp contact (not in list yet), shows in detail panel
- Click row → clones contact into temp, shows in detail panel
- Save → inserts temp (if new) or copies temp back to original (if editing)
- Cancel → discards temp, hides detail panel (original unchanged)
- Delete → removes original from list, clears detail
- No selection → hide detail panel (ui-class-hidden)
- Send chat → mcp.pushState({app="contacts", event="chat", text=...}) → parent responds via ui_run
