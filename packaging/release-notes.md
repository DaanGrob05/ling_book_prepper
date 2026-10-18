LingQ Prep prepares ebooks (EPUB) for importing into [LingQ](https://www.lingq.com): it repairs sentences that were cut over several lines and splits the book into lessons of at most 6000 words, one or more per chapter.

## Which file do I need?

| Your computer | Download |
|---|---|
| Linux (most PCs and laptops) | `LingQ-Prep-{{VERSION}}-Linux-x86_64.tar.xz` |

`SHA256SUMS.txt` is only needed to check that a download is complete. The **Source code** files are for developers; you don't need them to use the app.

## Installing

### Linux

Works on Ubuntu 22.04, Debian 12, Linux Mint 21, Fedora 36 or newer.

1. Download `LingQ-Prep-{{VERSION}}-Linux-x86_64.tar.xz`.
2. Right-click the file and choose **Extract Here**.
3. Open the extracted folder `LingQ-Prep-{{VERSION}}`. To try the app without installing it, double-click `lingqprep`.
4. To install it, right-click in an empty spot of the folder, choose **Open in Terminal**, type `./install.sh` and press Enter.

**LingQ Prep** is now in your app menu, and the extracted folder can be deleted. To remove the app again, run `./uninstall.sh` the same way. The folder's `README.txt` explains the same steps.
