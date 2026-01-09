Huntsman
========

Huntsman is a TUI app that spiders a website and lists all the resources it finds within that domain.

![Huntsman Screenshot](screenshot.png)

Installation
------------

### Quick Install (Recommended)

To install the latest version of `huntsman` (to `/usr/local/bin` on macOS or `~/.local/bin` on Linux):

```bash
curl -sL https://github.com/jturmel/huntsman/releases/latest/download/install.sh | bash
```

The script automatically detects your OS and architecture. Make sure the installation directory (`/usr/local/bin` on macOS or `~/.local/bin` on Linux) is in your `PATH`.

### Manual Install (From Source)

If you have Go installed, you can build and install manually:

```bash
git clone https://github.com/jturmel/huntsman.git
cd huntsman
make install
```

This will install the binary to `~/.local/bin/` (Linux) or `/usr/local/bin` (macOS), and a default `theme.json` to `~/.config/huntsman/` (Linux) or `~/Library/Application Support/huntsman/` (macOS).

Usage
-----

1. Run `huntsman`.
2. Enter the URL you want to spider in the input box.
3. Press **Enter** to start the crawl.
4. Use **Tab** to switch between the input box and the results table.
5. In the results table:
    - Use **Arrows** or **j/k** to scroll.
    - Press **/** to focus the filter input.
    - Advanced filtering:
        - By default, it filters by the **URL** column.
        - Use `type:{typevalue}` to filter by the **Type** column (e.g., `type:document`).
        - Use `status:{statusvalue}` to filter by the **Status** column (e.g., `status:404`).
    - Press **Enter** on a highlighted row to open the URL in your default browser.
    - Press **q** to quit.

Configuration
-------------

Huntsman supports custom color themes via a `theme.json` file. The app looks for this file in several locations (in order):
1. The current directory.
2. The same directory as the `huntsman` executable.
3. `~/.config/huntsman/theme.json`.
4. `~/Library/Application Support/huntsman/theme.json` (macOS only).

Example `theme.json`:

```json
{
  "focused_color": "#bd93f9",
  "blurred_color": "240",
  "spinner_color": "#bd93f9",
  "check_mark_color": "#bd93f9",
  "table_selected_fg": "229",
  "table_selected_bg": "#bd93f9"
}
```

License
-------

Huntsman is released under the [GNU GPLv3 License](LICENSE).
