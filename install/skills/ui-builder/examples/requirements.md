# Contact Manager with Chat

A contact management application that allows users to browse, search, and edit contacts. The interface includes an integrated chat panel for communicating with an AI agent.

## Features

### Contact List

The main view displays a scrollable list of contacts. Users can search contacts by typing in a search field, which filters the list in real-time. A badge next to the search field shows the count of matching contacts. Clicking a contact in the list selects it and displays its details in an adjacent panel.

### Contact Details

When a contact is selected, a detail panel appears showing editable fields for name, email, status (active or inactive), and a VIP toggle. Users can save changes, cancel edits, or delete the contact. The panel remains hidden when no contact is selected.

### Adding Contacts

An "Add" button creates a new empty contact and opens the detail panel for editing. The new contact only appears in the list after saving.

### Chat Panel

A chat interface at the bottom allows users to send messages to an AI agent. Messages appear in a scrollable history showing both user messages and agent responses. Sending a message notifies the parent application, which can respond with updates.

### Dark Mode

A toggle switch enables dark mode for the entire interface.

## Behavior Notes

- Search filtering happens as the user types
- Editing a contact works on a copy; changes only apply when saved
- Canceling discards all changes and returns the contact to its original state
- Deleting removes the contact immediately from the list
