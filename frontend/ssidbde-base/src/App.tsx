import { useState } from "react";
import type { Module } from "./type.d.ts";
import NavBar from "./NavBar.tsx";
import WelcomeModule from "./WelcomeModule.tsx";
import AboutModule from "./AboutModule.tsx";
import BackgroundModule from "./BackgroundModule.tsx";
import ProjectsModule from "./ProjectsModule.tsx";

function App() {

  const [diplayedModule, setDisplayedModule] = useState<Module>("welcome");

  let module;


  switch (diplayedModule) {
    case "welcome":
      module = <WelcomeModule />;
      break;
    case "about":
      module = <AboutModule />;
      break;
    case "background":
      module = <BackgroundModule />;
      break;
    case "projects":
      module = <ProjectsModule />;
      break;
  }

  return (
    <div className="h-full bg-linear-to-tl from-neutral-50  to-sky-100 text-zinc-800 flex flex-col">
      <div className="grow">
        {module}
      </div>
      <NavBar displayedModule={diplayedModule} setDisplayedModule={setDisplayedModule} />
    </div>
  )
}

export default App
