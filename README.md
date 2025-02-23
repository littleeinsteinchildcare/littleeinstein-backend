# Installing Go and Pre-Commit on Mac and Windows

## Installing Homebrew (Mac Only)
If you haven't installed Homebrew yet, you can find more details [here](https://brew.sh/).

1. Open Terminal.
2. Run the following command:
   ```sh
   /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
   ```
3. Add Homebrew to your PATH:
   ```sh
   echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
   eval "$(/opt/homebrew/bin/brew shellenv)"
   ```
4. Verify the installation:
   ```sh
   brew --version
   ```

---


## Installing Go

### macOS
#### Using Homebrew (Recommended)
1. Open the terminal and run:
   ```sh
   brew install go
   ```
2. Verify the installation:
   ```sh
   go version
   ```
3. Set up Go environment variables (optional but recommended):
   ```sh
   echo 'export GOPATH="$HOME/go"' >> ~/.zshrc
   echo 'export PATH="$GOPATH/bin:$PATH"' >> ~/.zshrc
   source ~/.zshrc
   ```
#### Manual Installation
1. Download the latest Go package from the official site: [Go Downloads](https://go.dev/dl/)
2. Open the downloaded `.pkg` file and follow the installation steps.
3. Verify the installation:
   ```sh
   go version
   ```

### Windows
#### Using Chocolatey (Recommended)
1. Open **PowerShell as Administrator** and run:
   ```powershell
   choco install golang -y
   ```
2. Restart your terminal and verify installation:
   ```powershell
   go version
   ```

#### Manual Installation
1. Download the Windows installer from [Go Downloads](https://go.dev/dl/).
2. Run the installer and follow the installation steps.
3. Ensure Go is added to your system `PATH` (usually done automatically).
4. Verify installation:
   ```powershell
   go version
   ```

## Installing Pre-Commit

### macOS
#### Using Homebrew (Recommended)
1. Install **pre-commit**:
   ```sh
   brew install pre-commit
   ```
2. Verify installation:
   ```sh
   pre-commit --version
   ```

#### Using Pip
1. Install **pre-commit** using Python's package manager:
   ```sh
   pip install pre-commit
   ```
2. Verify installation:
   ```sh
   pre-commit --version
   ```

### Windows
#### Using Chocolatey
1. Open **PowerShell as Administrator** and run:
   ```powershell
   choco install pre-commit -y
   ```
2. Verify installation:
   ```powershell
   pre-commit --version
   ```

#### Using Pip
1. Install **pre-commit** using Python's package manager:
   ```powershell
   pip install pre-commit
   ```
2. Verify installation:
   ```powershell
   pre-commit --version
   ```

## Setting Up Pre-Commit in a Repository
1. Navigate to your repository:
   ```sh
   cd $HOME/littleeinstein-backend
   ```
2. Installing commitlint/cli commitlint/config-conventional
   ```sh
   npm install -g @commitlint/cli @commitlint/config-conventional
   ```
   
3. Install pre-commit hooks:
   ```sh
   pre-commit install
   ```
4. Run pre-commit manually (optional, to test it):
   ```sh
   pre-commit run --all-files
   ```

Now your repository is set up with **pre-commit**, and hooks will run automatically on `git commit -m "Message"`!
