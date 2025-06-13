import CharacterVisualizer from "./CharacterVisualizer";

function AboutModule() {
  return (
    <div className="h-full flex items-center justify-center">
      <div className="w-[50%] flex gap-5 p-3 items-center h-full">
        <CharacterVisualizer />
        <div className="w-[50%]">
          Firefox detected an issue and did not continue to hugobde.dev. The website is either misconfigured or your computer clock is set to the wrong time.
          It’s likely the website’s certificate is expired, which prevents Firefox from connecting securely.
          What can you do about it?
          hugobde.dev has a security policy called HTTP Strict Transport Security (HSTS), which means that Firefox can only connect to it securely. You can’t add an exception to visit this site.
          The issue is most likely with the website, and there is nothing you can do to resolve it. You can notify the website’s administrator about the problem.
        </div>
      </div>
    </div>
  );
}

export default AboutModule;
