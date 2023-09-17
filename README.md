# SteamGuardDog

## Overview

SteamGuardDog is a utility tool designed to automate the process of logging into Steam via SteamCmd with multi-factor authentication (MFA) enabled. The tool fetches the MFA code from a Gmail account and uses it to proceed with the login, saving you the time and effort of having to manually enter the code yourself which can be especially useful for CI/CD pipelines or automated build setups.

## Why?

Manually entering MFA codes can be time-consuming, especially for CI/CD pipelines or automated build setups. This tool simplifies the login process by automating MFA code retrieval and input.

## Quick Start

For a hassle-free experience, pre-compiled binaries for SteamGuardDog are available on the [Releases page](https://github.com/your_username/SteamGuardDog/releases). Just download the appropriate version for your operating system and architecture, update `credentials.json` and `config.json` as outlined below. 

Upon the initial startup, a browser window will open and prompt you to authenticate with your Gmail account. Once you've done that, you can use SteamGuardDog just like you would use SteamCmd. For example:

```bash
./SteamGuardDog.exe +login username password +run_app_build /path/to/appconfig.vdf
```

## Installation Instructions

### Prerequisites
- [Go](https://golang.org/dl/) installed on your system (Version >= 1.16)
- [SteamCmd](https://developer.valvesoftware.com/wiki/SteamCMD) installed
- A Gmail account for receiving Steam Guard codes and the Gmail API setup correctly, instructions below
### Gmail API Setup

  - Visit the [Google Cloud Console](https://console.developers.google.com/)
  - Create a new project or use an existing one.
  - Navigate to **"APIs & Services" > "Credentials"**.
  - Create a new OAuth client ID and download the credentials.
  - During OAuth consent screen configuration, use External and add your build servers email as a test user
  - Save the downloaded `credentials.json` file in the same directory as your SteamGuardDog application.

  Make sure to enable the Gmail API for your project: Navigate to **"APIs & Services" > "Dashboard"**, click on **"+ ENABLE APIS AND SERVICES"**, search for "Gmail API" and enable it.

### Steps

1. **Clone the Repository**

    ```bash
    git clone https://github.com/yourusername/SteamGuardDog.git
    ```

2. **Navigate to the Project Directory**

    ```bash
    cd SteamGuardDog
    ```

3. **Build the Project**

    ```bash
    go build
    ```

    This will produce an executable named `SteamGuardDog`.

4. **Run the Tool**

    ```bash
    ./SteamGuardDog.exe +login username password [Your normal SteamCmd Arguments]
    ```

5. **(Optional) Move the executable to a global path**

    If you want to run `SteamGuardDog` from any directory, you can move the executable to a directory that's in your system's `PATH`.

    ```bash
    sudo mv ./SteamGuardDog /usr/local/bin/
    ```

## Usage

1. Configure the `config.json` file with the correct path to `steamcmd.exe` on your machine. Example:
    ```json
    {
        "steamcmd_path": "./steamcmd/steamcmd.exe"
    }
    ```

2. Make sure that `credentials.json` is correctly configured to use a GCP Project that you've made to authenticate with Gmail, as described in the [Gmail API Setup](#gmail-api-setup) section above.

2. You can use SteamGuardDog just like you would use SteamCmd. For example:

```bash
./SteamGuardDog.exe +login username password +run_app_build /path/to/appconfig.vdf +quit
```

## Contributing

If you'd like to contribute, please fork the repository and make changes as you'd like. Pull requests are warmly welcome.

## License

MIT © [Christian Casteel](https://christiancasteel.dev/)
