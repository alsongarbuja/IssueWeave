import { Marked } from "marked";
import DOMPurify from "isomorphic-dompurify";

const marked = new Marked({
  gfm: true,
  breaks: true,
});

export const useMarkdown = () => {
  const renderMarkdown = (rawBody: string) => {
    if (!rawBody) return "";

    const rawHTML = marked.parse(rawBody) as string;
    return DOMPurify.sanitize(rawHTML);
  };

  return {
    renderMarkdown,
  };
};
