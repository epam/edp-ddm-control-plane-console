export const getErrorMessage = (key: string): string => {
  switch (key) {
    case 'required':
      return 'Поле обов’язкове для заповнення.';
    case 'moreThanMaxValue':
      return 'Перевищено максимально допустиме значення';
    default:
      return 'Перевірте формат поля';
  }
};