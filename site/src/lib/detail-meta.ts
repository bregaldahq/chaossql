// Registers the titles of detail pages with route-meta. Imported only by the
// lazy page modules (and the prerender), so content JSON never reaches the
// main bundle.
import { docChapterFromPath, scenarioSlugFromPath } from './router';
import { metaDescription, registerMetaOverride } from './route-meta';
import { scenarioBySlug } from '../data/scenarios-data';
import { docChapter } from '../data/docs-content';

registerMetaOverride('scenarios', (pathname, lang) => {
  const scenario = scenarioBySlug(scenarioSlugFromPath(pathname));
  if (!scenario) return undefined;
  return {
    title: `${scenario.name[lang] || scenario.name.en} (${scenario.code}) | ${lang === 'pt' ? 'Cenários' : 'Scenarios'} | ChaosSQL`,
    description: metaDescription(scenario.summary[lang] || scenario.summary.en),
  };
});

registerMetaOverride('docs', (pathname, lang) => {
  const chapter = docChapter(docChapterFromPath(pathname), lang);
  if (!chapter) return undefined;
  return {
    title: `${chapter.title} | ${lang === 'pt' ? 'Documentação' : 'Docs'} | ChaosSQL`,
    description: metaDescription(chapter.summary),
  };
});
