/**
 * Portal action - moves element to a target container (default: body)
 *
 * Usage:
 * <div use:portal>Content rendered at body</div>
 * <div use:portal={'#modal-root'}>Content rendered at #modal-root</div>
 */

export interface PortalOptions {
  /** Target selector or element */
  target?: string | HTMLElement;
}

export function portal(node: HTMLElement, options: PortalOptions | string = {}) {
  let target: HTMLElement;

  function update(newOptions: PortalOptions | string = {}) {
    const opts = typeof newOptions === 'string' ? { target: newOptions } : newOptions;

    if (typeof opts.target === 'string') {
      target = document.querySelector(opts.target) as HTMLElement;
      if (!target) {
        console.warn(`Portal target "${opts.target}" not found, using body`);
        target = document.body;
      }
    } else if (opts.target instanceof HTMLElement) {
      target = opts.target;
    } else {
      target = document.body;
    }

    target.appendChild(node);
  }

  function destroy() {
    if (node.parentNode) {
      node.parentNode.removeChild(node);
    }
  }

  update(options);

  return {
    update,
    destroy,
  };
}

export default portal;
