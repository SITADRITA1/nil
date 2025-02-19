import { createDomain } from "effector";
import type { App } from "../../types";
import loadTutorials from "./spec";
export type Tutorial = {
  stage: number;
  text: string;
  contracts: string;
};

export const tutorialDomain = createDomain("tutorial");

export const $tutorial = tutorialDomain.createStore<Tutorial>({
  text: "",
  contracts: "",
  stage: 0,
});
export const $compiledTutorialContracts = tutorialDomain.createStore<App[]>([]);
export const $tutorialContracts = $tutorial.map((tutorial) => (tutorial ? tutorial.contracts : ""));

export const fetchTutorialEvent = tutorialDomain.createEvent<Tutorial>();
export const fetchAllTutorialsEvent = tutorialDomain.createEvent<Tutorial[]>();

export const fetchTutorialFx = tutorialDomain.createEffect<string, Tutorial, string>();
export const fetchAllTutorialsFx = tutorialDomain.createEffect<void, Tutorial[], string>();

fetchTutorialFx.use(async (stage) => {
  const tutorials = await loadTutorials();
  return tutorials.find((tutorial) => tutorial.stage === Number(stage))!;
});

fetchAllTutorialsFx.use(async () => {
  const tutorials = await loadTutorials();
  return tutorials;
});
