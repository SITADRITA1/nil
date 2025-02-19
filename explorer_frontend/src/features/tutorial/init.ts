import { sample } from "effector";
import { changeCode, loadedTutorialPage } from "../code/model";
import { tutorialWithStageRoute } from "../routing/routes/tutorialRoute";
import { $tutorial, fetchTutorialEvent, fetchTutorialFx } from "./model";

$tutorial.on(fetchTutorialFx.doneData, (_, tutorial) => tutorial);

sample({
  clock: loadedTutorialPage,
  source: tutorialWithStageRoute.$params,
  fn: (params) => params.stage,
  filter: (stage) => stage !== undefined,
  target: fetchTutorialFx,
});

sample({
  clock: fetchTutorialEvent,
  source: fetchTutorialEvent,
  fn: (tutorial) => tutorial.stage.toString(),
  target: fetchTutorialFx,
});

sample({
  clock: fetchTutorialFx.doneData,
  fn: (tutorial) => tutorial.contracts,
  target: changeCode,
});
