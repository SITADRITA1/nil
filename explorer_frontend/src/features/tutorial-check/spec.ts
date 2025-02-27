import { createDomain } from "effector";
import type { App } from "../../types";

export type TutorialCheck = {
  stage: number;
  check: () => Promise<boolean>;
}

export const tutorialCheckDomain = createDomain("tutorial-check");

export const $tutorialCheck = tutorialCheckDomain.createStore<TutorialCheck>({
  stage: 0,
  check: async () => true,
});

export const fetchTutorialCheckEvent = tutorialCheckDomain.createEvent<TutorialCheck>();

export const fetchTutorialCheckFx = tutorialCheckDomain.createEffect<number, TutorialCheck>();

fetchTutorialCheckFx.use(async (stage) => {
  const tutorialCheck = await import(`./checks/${stage}`);
  return tutorialCheck.default;
});