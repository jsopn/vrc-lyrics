# VRChat Lyrics
Display the track and its lyrics in VRChat's chatbox using the OSC.

<p align="center">
  <img src="https://media.jsopn.com/2023/06/rusk-supremacy.gif" alt="animated" />
</p>

## Installation
1. **Download** the latest version of the software from the [release section](https://github.com/jsopn/vrc-lyrics/releases).
2. **Copy** the `config.example.toml` file and rename it as `config.toml`. Place the `config.toml` file next to the software executable.
3. **Sign in** to your Spotify account by visiting [open.spotify.com](https://open.spotify.com) in your preferred web browser.
4. **Open devtools** in your web browser (shortcut: CTRL+SHIFT+I) and locate the request that contains the `sp_dc` cookie. Copy the value of the `sp_dc` cookie.
5. **Paste** the copied `sp_dc` cookie value into the `config.toml` file, in the appropriate field as indicated by the comments in the configuration file.
6. **Configure** the format settings in the `config.toml` file according to your preferences, following the instructions provided in the comments.
7. **Enable** the OSC (Open Sound Control) feature in VRChat.
8. **Start** the software by running the executable.
9. Enjoy displaying track information and lyrics in VRChat's chatbox!

## Software License
This software is released under the MIT License. You can find the license details in the [LICENSE](LICENSE) file.

## Contributing
Contributions to VRChat Lyrics are welcome! If you would like to contribute, please follow these steps:
1. Fork the repository on GitHub.
2. Make your changes or additions in a new branch.
3. Ensure that your code adheres to the project's coding conventions and style guide.
4. Test your changes thoroughly.
5. Commit your changes and push them to your forked repository.
6. Create a pull request on the main repository, describing your changes and why they should be merged.
7. Wait for the maintainers to review your pull request. Feedback or requests for further changes may be provided.
8. Once your pull request is approved, it will be merged into the main repository.

Thank you for contributing to VRChat Lyrics! Your help is greatly appreciated.
