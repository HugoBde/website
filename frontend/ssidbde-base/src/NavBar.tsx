import type { Module } from "./type";

const SELECTEDCLASS = " underline";

function NavBar({ displayedModule, setDisplayedModule }: { displayedModule: Module, setDisplayedModule(module: Module): void }) {

  const buttons = (["about", "background", "projects"] as Module[]).map(b => {
    if (b === displayedModule) {
      return (
        <button key={b} className={"text-white hover:cursor-pointer hover:underline underline-offset-3" + SELECTEDCLASS} onClick={() => setDisplayedModule(b)}>{b}</button>
      )
    } else {
      return (
        <button key={b} className="text-white hover:cursor-pointer hover:underline underline-offset-3" onClick={() => setDisplayedModule(b)}>{b}</button>
      )
    }
  })

  return (
    <div className="h-16 p-3 flex justify-end gap-3 bg-blackgrow bg-linear-(--navbar-gradient)">
      {buttons}
    </div >
  )
}

export default NavBar;
