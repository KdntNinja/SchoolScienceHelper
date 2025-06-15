# Anki Deck Import for Revision Feature

## Overview

This guide explains how to import a user's Anki decks into your platform's revision feature, allowing users to manage their cards and decks entirely within your site after import. There is no ongoing sync or communication with Anki after the initial import.

## Approach: One-Time Anki Deck Import

### Prerequisites

- User must have Anki desktop installed
- User must export their decks as `.apkg` files from Anki desktop

### Steps

1. **Export Deck from Anki**
   - Open Anki desktop
   - Select the deck you want to export
   - Go to `File > Export...`
   - Choose `Export format: Anki Deck Package (*.apkg)`
   - Save the `.apkg` file to your computer

2. **Import Deck into Your Site**
   - Go to the revision feature on your site
   - Use the provided import tool to upload the `.apkg` file
   - The site will parse the file and import all cards and deck structure

3. **Manage Decks and Cards on Your Site**
   - After import, all deck and card management is done within your platform
   - No changes are synced back to Anki
   - Users can edit, add, or delete cards and decks as needed

### Technical Notes

- Use an open-source `.apkg` parser (e.g., [anki-apkg-export](https://github.com/ospalh/anki-apkg-export) or [apkg-js](https://github.com/Arthur-Milchior/apkg-js)) to extract cards and deck info from the uploaded file
- Store imported decks/cards in your own database
- Do not attempt to write back to Anki or AnkiWeb

## Limitations

- No two-way sync: changes made on your site are not reflected in Anki
- Users must re-import if they want to update from Anki again
- Media (images/audio) in cards may require extra handling

## Security Notes

- Never ask for or store the user's AnkiWeb password
- All import is done client-side or via secure upload

## References

- [Anki Desktop](https://apps.ankiweb.net/)
- [anki-apkg-export (Python)](https://github.com/ospalh/anki-apkg-export)
- [apkg-js (JavaScript)](https://github.com/Arthur-Milchior/apkg-js)

---

This approach provides a simple, user-friendly way to migrate from Anki to your platform, with all future management handled on your site only.
