# Product Guidelines

## Tone and Style
- **Professional and Utilitarian:** Focus on clarity, efficiency, and information density.
- **Minimalist and Modern:** Prioritize clean aesthetics and a focused user experience.

## Visual Identity and TUI Design
- **Native Color Integration:** Prioritize the use of the user's terminal color scheme to ensure the TUI feels integrated with their environment.
- **Subtle Feedback:** Utilize Lip Gloss and Bubble Tea capabilities for smooth transitions and clear interactive states that provide feedback without being distracting.
- **Visual Status Indicators:** Use clear color coding or icons in the TUI to represent success, warning, and error states for discovered resources.

## User Experience (UX)
- **Keyboard-First Design:** Ensure all primary actions, including crawling, filtering, and navigation, are efficient and discoverable via keyboard shortcuts.
- **Informative Feedback:** Provide clear, non-intrusive feedback for long-running processes (like crawling) and immediate validation for user inputs.
- **Context-Aware Information:** Present detailed metadata or advanced options only when relevant to the current selection to maintain a clean interface.

## Codebase Standards
- **Modularity:** Maintain a clear separation of concerns between the crawling engine, TUI logic, and configuration management.
- **Testability:** Ensure core logic, particularly the crawler and filtering mechanisms, is covered by comprehensive unit tests.
- **Self-Documenting Code:** Adhere to clear naming conventions and keep functions focused and concise to ensure the codebase is easily understandable.
