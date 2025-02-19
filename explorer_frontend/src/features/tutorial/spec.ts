async function loadTutorials() {
  const testTutorial = await import("./assets/testTutorial.md?raw");
  const testContracts = await import("./assets/testContracts.sol?raw");
  const tutorials = [
    {
      stage: 1,
      text: testTutorial.default,
      contracts: testContracts.default,
    },
  ];
  return tutorials;
}

export default loadTutorials;
