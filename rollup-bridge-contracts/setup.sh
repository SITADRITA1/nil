#!/bin/bash

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check if .env file exists
if [ -f .env ]; then
    echo "Warning: .env file already exists. Skipping copy from .env.example."
else
    echo "Copying .env.example to .env..."
    cp .env.example .env
fi

echo "Installing npm dependencies..."
npm install

echo "Running Hardhat clean and compile..."
npx hardhat clean
npx hardhat compile

# Install Forge dependencies
if command_exists forge; then
    echo "Installing Forge dependencies..."

    # Check and install foundry-rs/forge-std if not already installed
    if [ ! -d "lib/forge-std" ]; then
        echo "Installing foundry-rs/forge-std..."
        forge install foundry-rs/forge-std --no-commit
    else
        echo "foundry-rs/forge-std already installed. Skipping."
    fi

    # Check and install transmissions11/solmate if not already installed
    if [ ! -d "lib/solmate" ]; then
        echo "Installing transmissions11/solmate..."
        forge install transmissions11/solmate --no-commit
    else
        echo "transmissions11/solmate already installed. Skipping."
    fi

    # Check and install dapphub/ds-test if not already installed
    if [ ! -d "lib/ds-test" ]; then
        echo "Installing dapphub/ds-test..."
        forge install dapphub/ds-test --no-commit
    else
        echo "dapphub/ds-test already installed. Skipping."
    fi

    # Remove .gitmodules file if it exists
    if [ -f .gitmodules ]; then
        echo "Removing .gitmodules file..."
        rm .gitmodules
    fi

    # Remove .gitmodules file if it exists
    if [ -f ../.gitmodules ]; then
        echo "Removing .gitmodules file..."
        rm ../.gitmodules
    fi

    echo "Running Forge clean and compile..."
    forge clean
    forge build

    echo "Running Forge tests..."
    forge test
else
    echo "Forge is not installed. Skipping Forge steps."
fi

# Health check
echo "Performing health check..."
if npx hardhat compile && forge build && forge test; then
    echo "Setup completed successfully!"
else
    echo "Setup failed. Please check the errors above."
    exit 1
fi
