async function loadTutorials() {
  const textOne = await import("./assets/tutorialOneText.md?raw");
  const contractsOne = await import("./assets/tutorialOneContracts.sol?raw");
  const tutorials = [
    {
      stage: 1,
      text: textOne.default,
      contracts: contractsOne.default,
    },
  ];
  return tutorials;
}

export default loadTutorials;
